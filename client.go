// Package aingle is a Go SDK for the AIngle Cortex REST API, the verifiable
// memory cortex for AI agents. It offers memory (remember, recall, semantic
// search), a semantic triple graph, and pattern queries over plain HTTP.
package aingle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the base URL used when none is supplied.
const DefaultBaseURL = "http://127.0.0.1:19090"

// Client is an HTTP client for the AIngle Cortex REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets the base URL of the AIngle Cortex server. If s is empty it is
// ignored and the default is retained.
func WithBaseURL(s string) Option {
	return func(c *Client) {
		if s != "" {
			c.baseURL = strings.TrimRight(s, "/")
		}
	}
}

// WithToken sets the bearer token sent in the Authorization header.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom *http.Client. This overrides any timeout set with
// WithTimeout unless WithTimeout is applied afterwards.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// NewClient creates a new AIngle Cortex client. Without options it targets
// DefaultBaseURL with a 30 second timeout and no auth token.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// do performs an HTTP request against the API and decodes a JSON response into
// out (when out is non-nil). A non-2xx status yields an *AIngleError.
func (c *Client) do(ctx context.Context, method, path string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("aingle: encode request: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("aingle: build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("aingle: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}

	if out == nil {
		// Drain the body so the connection can be reused.
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("aingle: decode response: %w", err)
	}
	return nil
}

// parseError reads a non-2xx response body and builds an *AIngleError.
func parseError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	message := strings.TrimSpace(string(data))

	// Prefer a structured message when the body is JSON.
	if len(data) > 0 {
		var envelope struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(data, &envelope); err == nil {
			if envelope.Message != "" {
				message = envelope.Message
			} else if envelope.Error != "" {
				message = envelope.Error
			}
		}
	}

	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}
	return &AIngleError{Status: resp.StatusCode, Message: message}
}

// Health returns the server health and subsystem statuses.
func (c *Client) Health(ctx context.Context) (*Health, error) {
	var out Health
	if err := c.do(ctx, http.MethodGet, "/api/v1/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stats returns graph and server statistics.
func (c *Client) Stats(ctx context.Context) (*Stats, error) {
	var out Stats
	if err := c.do(ctx, http.MethodGet, "/api/v1/stats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Remember stores a memory entry and returns its id.
func (c *Client) Remember(ctx context.Context, req RememberRequest) (*RememberResponse, error) {
	if req.Tags == nil {
		req.Tags = []string{}
	}
	var out RememberResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/memory/remember", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Recall retrieves memory entries matching the given filters.
func (c *Client) Recall(ctx context.Context, req RecallRequest) ([]RecallResult, error) {
	if req.Tags == nil {
		req.Tags = []string{}
	}
	var out []RecallResult
	if err := c.do(ctx, http.MethodPost, "/api/v1/memory/recall", req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Search performs a vector similarity search over stored memories.
func (c *Client) Search(ctx context.Context, req VectorSearchRequest) ([]RecallResult, error) {
	var out []RecallResult
	if err := c.do(ctx, http.MethodPost, "/api/v1/memory/search", req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MemoryStats returns statistics about short and long term memory.
func (c *Client) MemoryStats(ctx context.Context) (*MemoryStats, error) {
	var out MemoryStats
	if err := c.do(ctx, http.MethodGet, "/api/v1/memory/stats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Forget deletes a memory entry by id.
func (c *Client) Forget(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/memory/"+url.PathEscape(id), nil, nil)
}

// CreateTriple inserts a subject-predicate-object statement. The object is an
// untagged value: a string, number, bool, or a NodeRef for an IRI reference.
func (c *Client) CreateTriple(ctx context.Context, subject, predicate string, object interface{}) (*Triple, error) {
	req := CreateTripleRequest{Subject: subject, Predicate: predicate, Object: object}
	var out Triple
	if err := c.do(ctx, http.MethodPost, "/api/v1/triples", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTriples lists triples, optionally filtered by subject, predicate, object,
// limit, and offset.
func (c *Client) ListTriples(ctx context.Context, opts ListTriplesOptions) (*ListTriplesResponse, error) {
	q := url.Values{}
	if opts.Subject != "" {
		q.Set("subject", opts.Subject)
	}
	if opts.Predicate != "" {
		q.Set("predicate", opts.Predicate)
	}
	if opts.Object != "" {
		q.Set("object", opts.Object)
	}
	if opts.Limit != nil {
		q.Set("limit", strconv.Itoa(*opts.Limit))
	}
	if opts.Offset != nil {
		q.Set("offset", strconv.Itoa(*opts.Offset))
	}
	var out ListTriplesResponse
	if err := c.do(ctx, http.MethodGet, withQuery("/api/v1/triples", q), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTriple retrieves a single triple by id.
func (c *Client) GetTriple(ctx context.Context, id string) (*Triple, error) {
	var out Triple
	if err := c.do(ctx, http.MethodGet, "/api/v1/triples/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTriple deletes a triple by id.
func (c *Client) DeleteTriple(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/triples/"+url.PathEscape(id), nil, nil)
}

// Query matches triples against a pattern. Any of subject, predicate, or object
// may be left zero to act as a wildcard.
func (c *Client) Query(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	var out QueryResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/query", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Subjects lists distinct subjects, optionally filtered by predicate.
func (c *Client) Subjects(ctx context.Context, opts SubjectsOptions) (*SubjectsResponse, error) {
	q := url.Values{}
	if opts.Predicate != "" {
		q.Set("predicate", opts.Predicate)
	}
	if opts.Limit != nil {
		q.Set("limit", strconv.Itoa(*opts.Limit))
	}
	var out SubjectsResponse
	if err := c.do(ctx, http.MethodGet, withQuery("/api/v1/query/subjects", q), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Predicates lists distinct predicates, optionally filtered by subject.
func (c *Client) Predicates(ctx context.Context, opts PredicatesOptions) (*PredicatesResponse, error) {
	q := url.Values{}
	if opts.Subject != "" {
		q.Set("subject", opts.Subject)
	}
	if opts.Limit != nil {
		q.Set("limit", strconv.Itoa(*opts.Limit))
	}
	var out PredicatesResponse
	if err := c.do(ctx, http.MethodGet, withQuery("/api/v1/query/predicates", q), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// withQuery appends encoded query parameters to a path when any are present.
func withQuery(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}
