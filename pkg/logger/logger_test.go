package logger

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLogEntryJSON(t *testing.T) {
	entry := LogEntry{
		Severity:   "ERROR",
		Message:    "request failed",
		Service:    "payments",
		Timestamp:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Attributes: map[string]any{"request_id": "abc"},
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LogEntry
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Severity != entry.Severity || decoded.Message != entry.Message ||
		decoded.Service != entry.Service || !decoded.Timestamp.Equal(entry.Timestamp) ||
		decoded.Attributes["request_id"] != entry.Attributes["request_id"] {
		t.Fatalf("round trip mismatch: got %#v, want %#v", decoded, entry)
	}
}
