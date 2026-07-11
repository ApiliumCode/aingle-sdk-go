# AIngle SDK for Go

Go SDK for [AIngle](https://apilium.com), the verifiable memory cortex for AI agents.

AIngle Cortex gives an agent durable memory and a semantic graph over plain HTTP.
You can remember facts, recall them by text or tags, run vector similarity search,
and store subject-predicate-object triples that you can query by pattern. The SDK
uses only the Go standard library.

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
	// Defaults to http://127.0.0.1:19090. Add options as needed.
	client := aingle.NewClient()

	ctx := context.Background()

	// Remember a fact.
	remembered, err := client.Remember(ctx, aingle.RememberRequest{
		EntryType:  "note",
		Data:       map[string]any{"text": "The launch is on Friday."},
		Tags:       []string{"launch", "schedule"},
		Importance: 0.8,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("remembered %s\n", remembered.ID)

	// Recall it by text.
	results, err := client.Recall(ctx, aingle.RecallRequest{
		Text: "when is the launch",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range results {
		fmt.Printf("[%.2f] %s: %v\n", r.Relevance, r.EntryType, r.Data)
	}
}
```

## Configuration

`NewClient` takes functional options. All are optional.

```go
client := aingle.NewClient(
	aingle.WithBaseURL("http://127.0.0.1:19090"),
	aingle.WithToken("my-namespace-token"),
	aingle.WithTimeout(15*time.Second),
)
```

## Triples and the semantic graph

A triple is a subject, a predicate, and an object. The object is an untagged
value: a string, number, boolean, or an IRI node reference. Use `NodeRef` to
build a node reference, which serializes to `{"node": "..."}`.

```go
// Object as a string literal.
_, err := client.CreateTriple(ctx, "alice", "knows", "bob")

// Object as a node reference (IRI).
_, err = client.CreateTriple(ctx, "alice", "homepage",
	aingle.NodeRef{Node: "http://example.org/alice"})

// Query by pattern. Leave a field empty to treat it as a wildcard.
res, err := client.Query(ctx, aingle.QueryRequest{Subject: "alice"})
for _, t := range res.Matches {
	fmt.Printf("%s %s %v\n", t.Subject, t.Predicate, t.Object)
}
```

On output, a triple's `Object` decodes into `interface{}`. A node reference
decodes into a `map[string]any` with a `node` key.

## Error handling

Non-2xx responses return a typed `*AIngleError` carrying the HTTP status and a
message.

```go
_, err := client.GetTriple(ctx, "missing-id")
var apiErr *aingle.AIngleError
if errors.As(err, &apiErr) {
	fmt.Printf("status %d: %s\n", apiErr.Status, apiErr.Message)
}
```

## API reference

| Method | Description |
|--------|-------------|
| `Health(ctx)` | Server health and subsystem status |
| `Stats(ctx)` | Graph and server statistics |
| `Remember(ctx, req)` | Store a memory entry, returns its id |
| `Recall(ctx, req)` | Retrieve memories by text, tags, or type |
| `Search(ctx, req)` | Vector similarity search over memories |
| `MemoryStats(ctx)` | Short and long term memory statistics |
| `Forget(ctx, id)` | Delete a memory entry |
| `CreateTriple(ctx, subject, predicate, object)` | Insert a triple |
| `ListTriples(ctx, opts)` | List triples with optional filters |
| `GetTriple(ctx, id)` | Fetch a triple by id |
| `DeleteTriple(ctx, id)` | Delete a triple by id |
| `Query(ctx, req)` | Match triples against a pattern |
| `Subjects(ctx, opts)` | List distinct subjects |
| `Predicates(ctx, opts)` | List distinct predicates |

## License

Apache-2.0, see [LICENSE](LICENSE).

## Links

- [AIngle Core](https://github.com/ApiliumCode/aingle)
- [Documentation](https://docs.apilium.com)
