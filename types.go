package aingle

import "fmt"

// Version is the SDK version.
const Version = "0.2.0"

// AIngleError is a typed error returned when the AIngle Cortex API responds with
// a non-2xx status. It carries the HTTP status code and a human-readable message.
type AIngleError struct {
	// Status is the HTTP status code returned by the server.
	Status int
	// Message is the error message, taken from the response body when available.
	Message string
}

func (e *AIngleError) Error() string {
	return fmt.Sprintf("aingle: status %d: %s", e.Status, e.Message)
}

// NodeRef is a helper for constructing a triple object that is an IRI node
// reference. It marshals to {"node": "..."}.
type NodeRef struct {
	Node string `json:"node"`
}

// Component describes the status of a single server subsystem in a health check.
type Component struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthComponents groups the individual subsystem health entries.
type HealthComponents struct {
	Graph Component `json:"graph"`
	Logic Component `json:"logic"`
}

// Health is the response from GET /api/v1/health.
type Health struct {
	Status     string           `json:"status"`
	Components HealthComponents `json:"components"`
}

// GraphStats holds counts describing the semantic graph.
type GraphStats struct {
	TripleCount    int64 `json:"triple_count"`
	SubjectCount   int64 `json:"subject_count"`
	PredicateCount int64 `json:"predicate_count"`
	ObjectCount    int64 `json:"object_count"`
}

// ServerStats holds runtime information about the server.
type ServerStats struct {
	ConnectedClients int    `json:"connected_clients"`
	UptimeSeconds    int64  `json:"uptime_seconds"`
	Version          string `json:"version"`
}

// Stats is the response from GET /api/v1/stats.
type Stats struct {
	Graph  GraphStats  `json:"graph"`
	Server ServerStats `json:"server"`
}

// RememberRequest is the body for POST /api/v1/memory/remember.
type RememberRequest struct {
	EntryType  string      `json:"entry_type"`
	Data       interface{} `json:"data"`
	Tags       []string    `json:"tags"`
	Importance float64     `json:"importance"`
	Embedding  []float64   `json:"embedding,omitempty"`
}

// RememberResponse is the response from POST /api/v1/memory/remember.
type RememberResponse struct {
	ID string `json:"id"`
}

// RecallRequest is the body for POST /api/v1/memory/recall.
type RecallRequest struct {
	Text          string   `json:"text,omitempty"`
	Tags          []string `json:"tags"`
	EntryType     string   `json:"entry_type,omitempty"`
	MinImportance *float64 `json:"min_importance,omitempty"`
	Limit         *int     `json:"limit,omitempty"`
}

// RecallResult is a single memory entry returned by recall and search.
type RecallResult struct {
	ID           string      `json:"id"`
	EntryType    string      `json:"entry_type"`
	Data         interface{} `json:"data"`
	Tags         []string    `json:"tags"`
	Importance   float64     `json:"importance"`
	Relevance    float64     `json:"relevance"`
	Source       string      `json:"source"`
	CreatedAt    string      `json:"created_at"`
	LastAccessed string      `json:"last_accessed"`
	AccessCount  int64       `json:"access_count"`
}

// VectorSearchRequest is the body for POST /api/v1/memory/search.
type VectorSearchRequest struct {
	Embedding     []float64 `json:"embedding"`
	K             int       `json:"k"`
	MinSimilarity float64   `json:"min_similarity"`
	EntryType     string    `json:"entry_type,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
}

// MemoryStats is the response from GET /api/v1/memory/stats.
type MemoryStats struct {
	STMCount         int64 `json:"stm_count"`
	STMCapacity      int64 `json:"stm_capacity"`
	LTMEntityCount   int64 `json:"ltm_entity_count"`
	LTMLinkCount     int64 `json:"ltm_link_count"`
	TotalMemoryBytes int64 `json:"total_memory_bytes"`
}

// Triple is a subject-predicate-object statement in the semantic graph. The
// Object is an untagged union decoded into interface{} (string, number, bool,
// or a map for a node reference).
type Triple struct {
	ID        string      `json:"id,omitempty"`
	Subject   string      `json:"subject"`
	Predicate string      `json:"predicate"`
	Object    interface{} `json:"object"`
	CreatedAt string      `json:"created_at,omitempty"`
}

// CreateTripleRequest is the body for POST /api/v1/triples.
type CreateTripleRequest struct {
	Subject   string      `json:"subject"`
	Predicate string      `json:"predicate"`
	Object    interface{} `json:"object"`
}

// ListTriplesOptions holds the optional filters for listing triples.
type ListTriplesOptions struct {
	Subject   string
	Predicate string
	Object    string
	Limit     *int
	Offset    *int
}

// ListTriplesResponse is the response from GET /api/v1/triples.
type ListTriplesResponse struct {
	Triples []Triple `json:"triples"`
	Total   int64    `json:"total"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
}

// QueryRequest is the body for POST /api/v1/query.
type QueryRequest struct {
	Subject   string      `json:"subject,omitempty"`
	Predicate string      `json:"predicate,omitempty"`
	Object    interface{} `json:"object,omitempty"`
	Limit     *int        `json:"limit,omitempty"`
}

// QueryResponse is the response from POST /api/v1/query.
type QueryResponse struct {
	Matches []Triple    `json:"matches"`
	Total   int64       `json:"total"`
	Pattern interface{} `json:"pattern"`
}

// SubjectsOptions holds the optional filters for listing subjects.
type SubjectsOptions struct {
	Predicate string
	Limit     *int
}

// SubjectsResponse is the response from GET /api/v1/query/subjects.
type SubjectsResponse struct {
	Subjects []string `json:"subjects"`
	Total    int64    `json:"total"`
}

// PredicatesOptions holds the optional filters for listing predicates.
type PredicatesOptions struct {
	Subject string
	Limit   *int
}

// PredicatesResponse is the response from GET /api/v1/query/predicates.
type PredicatesResponse struct {
	Predicates []string `json:"predicates"`
	Total      int64    `json:"total"`
}
