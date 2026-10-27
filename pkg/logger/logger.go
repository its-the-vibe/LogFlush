package logger

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// LogEntry is the structured log payload shared by producers and the worker.
type LogEntry struct {
	Severity   string         `json:"severity"`
	Message    string         `json:"message"`
	Service    string         `json:"service"`
	Timestamp  time.Time      `json:"timestamp"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Client pushes structured log entries onto a Redis list.
type Client struct {
	redis redis.UniversalClient
	key   string
}

// New creates a logger that pushes entries to key using client.
func New(client redis.UniversalClient, key string) (*Client, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if key == "" {
		return nil, errors.New("redis key is required")
	}
	return &Client{redis: client, key: key}, nil
}

// Log serializes entry as JSON and pushes it onto the configured Redis list.
func (c *Client) Log(ctx context.Context, entry LogEntry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return c.redis.LPush(ctx, c.key, string(payload)).Err()
}
