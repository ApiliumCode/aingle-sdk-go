// Package aingle provides a Go SDK for interacting with AIngle nodes.
package aingle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ClientConfig holds configuration for the AIngle client.
type ClientConfig struct {
	// NodeURL is the HTTP URL of the AIngle node
	NodeURL string
	// WSURL is the WebSocket URL for real-time updates
	WSURL string
	// Timeout is the request timeout
	Timeout time.Duration
	// Debug enables debug logging
	Debug bool
}

// DefaultConfig returns the default client configuration.
func DefaultConfig() ClientConfig {
	return ClientConfig{
		NodeURL: "http://localhost:8080",
		WSURL:   "ws://localhost:8081",
		Timeout: 30 * time.Second,
		Debug:   false,
	}
}

// Client is the main entry point for interacting with AIngle nodes.
type Client struct {
	config     ClientConfig
	httpClient *http.Client
	wsConn     *websocket.Conn
}

// NewClient creates a new AIngle client with the given configuration.
func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// NewDefaultClient creates a new AIngle client with default configuration.
func NewDefaultClient() *Client {
	return NewClient(DefaultConfig())
}

// CreateEntry creates a new entry in the DAG.
func (c *Client) CreateEntry(ctx context.Context, data interface{}) (EntryHash, error) {
	payload := map[string]interface{}{
		"data": data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.NodeURL+"/api/v1/entries", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Hash EntryHash `json:"hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Hash, nil
}

// GetEntry retrieves an entry by hash.
func (c *Client) GetEntry(ctx context.Context, hash EntryHash) (*Entry, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.NodeURL+"/api/v1/entries/"+string(hash), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var entry Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &entry, nil
}

// GetNodeInfo retrieves node information.
func (c *Client) GetNodeInfo(ctx context.Context) (*NodeInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.NodeURL+"/api/v1/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var info NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &info, nil
}

// Subscribe subscribes to real-time entry updates.
// Returns a channel that receives entries and a function to unsubscribe.
func (c *Client) Subscribe(ctx context.Context) (<-chan *Entry, func(), error) {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.config.WSURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.wsConn = conn
	entryChan := make(chan *Entry, 100)

	go func() {
		defer close(entryChan)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var entry Entry
			if err := json.Unmarshal(message, &entry); err != nil {
				continue
			}

			select {
			case entryChan <- &entry:
			default:
				// Channel full, skip
			}
		}
	}()

	unsubscribe := func() {
		conn.Close()
	}

	return entryChan, unsubscribe, nil
}

// Close closes the client and any open connections.
func (c *Client) Close() error {
	if c.wsConn != nil {
		return c.wsConn.Close()
	}
	return nil
}
