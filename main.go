package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Server struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Response struct {
	Index    int     `json:"index"`
	Type     string  `json:"type"`
	Latency  float64 `json:"latency_ms"`
	ServerID int     `json:"server_id"`
}

type ServerResult struct {
	ServerID      int        `json:"server_id"`
	ServerAddress string     `json:"server_address"`
	Responses     []Response `json:"responses"`
}

type AddressReport struct {
	Address      string         `json:"address"`
	MinLatency   float64       `json:"min_latency_ms"`
	MaxLatency   float64       `json:"max_latency_ms"`
	AvgLatency   float64       `json:"avg_latency_ms"`
	TimeoutCount int           `json:"timeout_count"`
	OnlineCount  int           `json:"online_count"`
	OfflineCount int           `json:"offline_count"`
	TotalCount   int           `json:"total_count"`
	Servers      []ServerResult `json:"servers"`
}

const (
	timeoutDuration = 6 * time.Second
	retryCount     = 3
)

func pingServer(serverID int, serverAddr string, targetAddr string, index int) Response {
	url := fmt.Sprintf("http://%s/PING/%s?trID=%d&nPing=1", serverAddr, targetAddr, index)
	
	client := &http.Client{
		Timeout: timeoutDuration,
	}
	
	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start).Milliseconds()
	
	if err != nil {
		return Response{
			Index:    index,
			Type:     "server-offline",
			Latency:  -1,
			ServerID: serverID,
		}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return Response{
			Index:    index,
			Type:     "timeout",
			Latency:  float64(latency),
			ServerID: serverID,
		}
	}
	
	return Response{
		Index:    index,
		Type:     "online",
		Latency:  float64(latency),
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