package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
    "sync"
	"time"
)

const (
    retryCount = 3
    timeout  = 5 * time.Second
)

type Server struct {
    ID      int    `json:"id"`
    Address string `json:"address"`
}

type ServerResult struct {
    ServerID      int        `json:"server_id"`
    ServerAddress string     `json:"server_address"`
    Responses     []Response `json:"responses"`
}

type AddressReport struct {
    Address      string         `json:"address"`
    MinLatency   float64       `json:"min_latency"`
    MaxLatency   float64       `json:"max_latency"`
    AvgLatency   float64       `json:"avg_latency"`
    TimeoutCount int           `json:"timeout_count"`
    OnlineCount  int           `json:"online_count"`
    OfflineCount int           `json:"offline_count"`
    TotalCount   int           `json:"total_count"`
    Servers      []ServerResult `json:"servers"`
}

type PingResponse struct {
    Datetime string `json:"datetime"`
    Err      *struct {
        Message string `json:"message"`
        Name    string `json:"name"`
    } `json:"err"`
    Ms     int64  `json:"ms"`
    Query  Query  `json:"query"`
    SID    int64  `json:"sID"`
    Target string `json:"target"`
    TTL    int    `json:"ttl"`
}

type Query struct {
    NPing string `json:"nPing"`
    TrID  string `json:"trID"`
}

type Response struct {
    Index    int     `json:"index"`
    Type     string  `json:"type"`
    Latency  float64 `json:"latency"`
    ServerID int     `json:"server_id"`
}


func pingServer(serverID int, serverAddr string, targetAddr string, index int) Response {
    // Construct the ping URL
    url := fmt.Sprintf("http://%s/PING/%s?trID=%d&nPing=1", serverAddr, targetAddr, index)
    
    // Initialize HTTP client
    client := &http.Client{
		Timeout: timeout,
	}
    
    // Create and execute the request
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return Response{
            Index:    index,
            Type:     "server-offline",
            Latency:  -1,
            ServerID: serverID,
        }
    }
    
    // Execute the request
    resp, err := client.Do(req)
    if err != nil {
        return Response{
            Index:    index,
            Type:     "server-offline",
            Latency:  -1,
            ServerID: serverID,
        }
    }
    defer resp.Body.Close()
    
    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return Response{
            Index:    index,
            Type:     "server-offline",
            Latency:  -1,
            ServerID: serverID,
        }
    }
    
    // Parse JSON response
    var pingResp PingResponse
    if err := json.Unmarshal(body, &pingResp); err != nil {
        return Response{
            Index:    index,
            Type:     "server-offline",
            Latency:  -1,
            ServerID: serverID,
        }
    }
    
    // Handle response based on error presence
    if pingResp.Err != nil {
        return Response{
            Index:    index,
            Type:     "timeout",
            Latency:  float64(pingResp.Ms),
            ServerID: serverID,
        }
    }
    
    // Return successful response
    return Response{
        Index:    index,
        Type:     "online",
        Latency:  float64(pingResp.Ms),
        ServerID: serverID,
    }
}

func testServer(server Server, targetAddr string) ServerResult {
	var responses []Response
	
	for i := 0; i < retryCount; i++ {
		response := pingServer(server.ID, server.Address, targetAddr, i)
		responses = append(responses, response)
	}
	
	return ServerResult{
		ServerID:      server.ID,
		ServerAddress: server.Address,
		Responses:     responses,
	}
}

func calculateStats(results []ServerResult) (float64, float64, float64, int, int, int, int) {
	var (
		minLatency    float64 = -1
		maxLatency    float64
		totalLatency  float64
		onlineCount   int
		offlineCount  int
		timeoutCount  int
		totalCount    int
	)
	
	for _, server := range results {
		for _, resp := range server.Responses {
			totalCount++
			
			switch resp.Type {
			case "online":
				onlineCount++
				totalLatency += resp.Latency
				
				if minLatency == -1 || resp.Latency < minLatency {
					minLatency = resp.Latency
				}
				if resp.Latency > maxLatency {
					maxLatency = resp.Latency
				}
				
			case "timeout":
				timeoutCount++
			case "server-offline":
				offlineCount++
			}
		}
	}
	
	var avgLatency float64
	if onlineCount > 0 {
		avgLatency = totalLatency / float64(onlineCount)
	}
	
	if minLatency == -1 {
		minLatency = 0
	}
	
	return minLatency, maxLatency, avgLatency, timeoutCount, onlineCount, offlineCount, totalCount
}

func testAddress(address string, servers []Server) AddressReport {
	var (
		wg      sync.WaitGroup
		results = make([]ServerResult, len(servers))
	)
	
	for i, server := range servers {
		wg.Add(1)
		go func(idx int, srv Server) {
			defer wg.Done()
			results[idx] = testServer(srv, address)
		}(i, server)
	}
	
	wg.Wait()
	
	minLatency, maxLatency, avgLatency, timeoutCount, onlineCount, offlineCount, totalCount := calculateStats(results)
	
	return AddressReport{
		Address:      address,
		MinLatency:   float64(int(minLatency*100)) / 100,
		MaxLatency:   float64(int(maxLatency*100)) / 100,
		AvgLatency:   float64(int(avgLatency*100)) / 100,
		TimeoutCount: timeoutCount,
		OnlineCount:  onlineCount,
		OfflineCount: offlineCount,
		TotalCount:   totalCount,
		Servers:      results,
	}
}

func main() {
	serversFile := flag.String("servers", "/usr/lib/zabbix/externalscripts/servers.json", "Path to servers JSON file")
	flag.Parse()
	
	addresses := flag.Args()
	if len(addresses) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No addresses provided")
		fmt.Fprintln(os.Stderr, "Usage: program -servers=servers.json address1 [address2 ...]")
		os.Exit(1)
	}
	
	serversData, err := os.ReadFile(*serversFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading servers file: %v\n", err)
		os.Exit(1)
	}
	
	var servers []Server
	if err := json.Unmarshal(serversData, &servers); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing servers JSON: %v\n", err)
		os.Exit(1)
	}
	
	var reports []AddressReport
	for _, addr := range addresses {
		report := testAddress(addr, servers)
		reports = append(reports, report)
	}
	
	output, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(output))
}