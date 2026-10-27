package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/its-the-vibe/LogFlush/pkg/logger"
)

type testWriter struct {
	failures int
	calls    int
}

func (w *testWriter) WriteBatch(_ context.Context, entries []logger.LogEntry) error {
	w.calls++
	if len(entries) != 1 {
		return errors.New("unexpected batch size")
	}
	if w.calls <= w.failures {
		return errors.New("temporary failure")
	}
	return nil
}

func TestFlushRetriesAndClearsBatch(t *testing.T) {
	writer := &testWriter{failures: 1}
	worker := &Worker{
		writer: writer,
		batch:  []logger.LogEntry{{Message: "hello"}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := worker.flushWithRetry(ctx); err != nil {
		t.Fatal(err)
	}
	if writer.calls != 2 {
		t.Fatalf("writer called %d times, want 2", writer.calls)
	}
	if len(worker.batch) != 0 {
		t.Fatalf("batch not cleared after successful write: %#v", worker.batch)
	}
}

func TestFlushPreservesBatchWhenContextEnds(t *testing.T) {
	worker := &Worker{
		writer: &testWriter{failures: 1},
		batch:  []logger.LogEntry{{Message: "hello"}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := worker.flushWithRetry(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
	if len(worker.batch) != 1 {
		t.Fatal("batch should remain available when writing fails")
	}
}
