# Network Monitor Tool

A Go-based network monitoring tool that performs latency tests against multiple servers and generates detailed JSON reports.

## Features

- Multi-server latency testing
- Concurrent server checks
- Configurable retry attempts
- JSON output format
- Detailed statistics per address

## Installation

1. Clone the repository or create the following directory structure:
```
network-monitor/
├── main.go
└── servers.json
```

2. Build the tool:
```bash
go build -o network-monitor main.go
```

## Configuration

### Server Configuration (servers.json)

Create a `servers.json` file with your server list:

```json
[
  {
    "id": 1,
    "name": "Server 1",
    "address": "server1.example.com"
  },
  {
    "id": 2,
    "name": "Server 2",
    "address": "server2.example.com"
  }
]
```

### Constants

The following constants can be modified in the code:
- `timeoutDuration`: HTTP request timeout (default: 6 seconds)
- `retryCount`: Number of ping attempts per server (default: 3)

## Usage

### Basic Usage

```bash
./network-monitor 8.8.8.8 1.1.1.1
```

### Custom Server File

```bash
./network-monitor -servers=custom_servers.json 8.8.8.8
```

## Output Format

The tool outputs JSON-formatted results:

```json
[
  {
    "address": "1.1.1.1",
    "min_latency_ms": 15.24,
    "max_latency_ms": 45.67,
    "avg_latency_ms": 30.45,
    "timeout_count": 1,
    "online_count": 5,
    "offline_count": 0,
    "total_count": 6,
    "servers": [
      {
        "server_id": 1,
        "server_address": "server1.example.com",
        "responses": [
          {
            "index": 0,
            "type": "online",
            "latency_ms": 15.24,
            "server_id": 1
          }
        ]
      }
    ]
  }
]
```

### Response Types

- `online`: Server responded successfully
- `timeout`: Server response exceeded timeout duration
- `server-offline`: Test server could not be reached

## Troubleshooting

### Common Issues

1. "Error parsing servers JSON":
   - Verify servers.json format matches the example above
   - Check JSON syntax

2. "No addresses provided":
   - Ensure at least one IP address is provided as argument

3. "Error reading servers file":
   - Verify servers.json exists in the correct location
   - Check file permissions

## License

This project is licensed under the MIT License - see the LICENSE file for details.