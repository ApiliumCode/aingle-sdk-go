# AIngle SDK for Go

Official Go SDK for [AIngle](https://apilium.com) - the ultra-light distributed ledger for IoT devices.

## Installation

```bash
go get github.com/ApiliumCode/aingle-sdk-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	aingle "github.com/ApiliumCode/aingle-sdk-go"
)

func main() {
	// Create client with default config
	client := aingle.NewDefaultClient()
	defer client.Close()

	ctx := context.Background()

	// Create an entry
	hash, err := client.CreateEntry(ctx, map[string]interface{}{
		"type":  "sensor_reading",
		"value": 23.5,
		"unit":  "celsius",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created entry: %s\n", hash)

	// Retrieve an entry
	entry, err := client.GetEntry(ctx, hash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Entry: %+v\n", entry)

	// Get node info
	info, err := client.GetNodeInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node version: %s\n", info.Version)
}
```

## Subscribe to Real-time Updates

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	aingle "github.com/ApiliumCode/aingle-sdk-go"
)

func main() {
	client := aingle.NewDefaultClient()
	defer client.Close()

	ctx := context.Background()

	entries, unsubscribe, err := client.Subscribe(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	// Listen for 60 seconds
	timeout := time.After(60 * time.Second)
	for {
		select {
		case entry := <-entries:
			fmt.Printf("New entry: %s\n", entry.Hash)
		case <-timeout:
			return
		}
	}
}
```

## API Reference

### Client

| Method | Description |
|--------|-------------|
| `NewClient(config)` | Create client with config |
| `NewDefaultClient()` | Create client with defaults |
| `CreateEntry(ctx, data)` | Create a new entry |
| `GetEntry(ctx, hash)` | Retrieve an entry by hash |
| `GetNodeInfo(ctx)` | Get node information |
| `Subscribe(ctx)` | Subscribe to real-time updates |
| `Close()` | Close the client |

### Configuration

```go
config := aingle.ClientConfig{
    NodeURL: "http://localhost:8080",
    WSURL:   "ws://localhost:8081",
    Timeout: 30 * time.Second,
    Debug:   false,
}
client := aingle.NewClient(config)
```

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter
golangci-lint run
```

## License

Apache-2.0 - see [LICENSE](LICENSE)

## Links

- [AIngle Core](https://github.com/ApiliumCode/aingle)
- [Documentation](https://docs.apilium.com)
- [Discord](https://discord.gg/apilium)
