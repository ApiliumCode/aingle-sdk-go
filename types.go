package aingle

import "fmt"

// EntryHash is a Blake3 hash as hex string.
type EntryHash string

// AgentPubKey is an Ed25519 public key as hex string.
type AgentPubKey string

// Entry represents an entry in the AIngle DAG.
type Entry struct {
	Hash      EntryHash    `json:"hash"`
	Author    AgentPubKey  `json:"author"`
	Parents   []EntryHash  `json:"parents"`
	Data      interface{}  `json:"data"`
	Timestamp int64        `json:"timestamp"`
	Sequence  uint32       `json:"sequence"`
	Signature string       `json:"signature"`
}

// NodeInfo contains information about an AIngle node.
type NodeInfo struct {
	NodeID         string   `json:"node_id"`
	Version        string   `json:"version"`
	Uptime         int64    `json:"uptime"`
	EntriesCount   int64    `json:"entries_count"`
	PeersCount     int      `json:"peers_count"`
	StorageBackend string   `json:"storage_backend"`
	Features       []string `json:"features"`
}

// PeerInfo contains information about a connected peer.
type PeerInfo struct {
	PeerID    string `json:"peer_id"`
	Address   string `json:"address"`
	Quality   int    `json:"quality"`
	LastSeen  int64  `json:"last_seen"`
	LatestSeq uint32 `json:"latest_seq"`
}

// SyncStatus represents the synchronization status.
type SyncStatus struct {
	Syncing  bool  `json:"syncing"`
	Pending  int   `json:"pending"`
	LastSync int64 `json:"last_sync"`
}

// ErrorCode represents error types.
type ErrorCode string

const (
	ErrConnectionFailed ErrorCode = "CONNECTION_FAILED"
	ErrTimeout          ErrorCode = "TIMEOUT"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrInvalidEntry     ErrorCode = "INVALID_ENTRY"
	ErrStorage          ErrorCode = "STORAGE_ERROR"
	ErrNetwork          ErrorCode = "NETWORK_ERROR"
	ErrAuth             ErrorCode = "AUTH_ERROR"
)

// AIngleError is an SDK error.
type AIngleError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *AIngleError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AIngleError) Unwrap() error {
	return e.Cause
}
