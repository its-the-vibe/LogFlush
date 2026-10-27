package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/its-the-vibe/LogFlush/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	redisPollInterval = time.Second
	maxRetryDelay     = 30 * time.Second
	shutdownTimeout   = 30 * time.Second
)

type Worker struct {
	redis         redis.UniversalClient
	key           string
	writer        logWriter
	batchSize     int
	flushInterval time.Duration
	batch         []logger.LogEntry
}

// New creates a worker that reads from key and writes batches to writer.
func New(client redis.UniversalClient, key string, writer logWriter, batchSize int, flushInterval time.Duration) (*Worker, error) {
	if client == nil || writer == nil {
		return nil, errors.New("redis client and log writer are required")
	}
	if key == "" || batchSize < 1 || flushInterval <= 0 {
		return nil, errors.New("key, positive batch size, and positive flush interval are required")
	}
	return &Worker{
		redis:         client,
		key:           key,
		writer:        writer,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		batch:         make([]logger.LogEntry, 0, batchSize),
	}, nil
}

// Run consumes Redis entries until ctx is canceled, then attempts a final flush.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		if len(w.batch) >= w.batchSize {
			if err := w.flushWithRetry(ctx); err != nil {
				return err
			}
		}

		select {
		case <-ctx.Done():
			flushCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			return w.flushWithRetry(flushCtx)
		case <-ticker.C:
			if err := w.flushWithRetry(ctx); err != nil && ctx.Err() == nil {
				log.Printf("flush logs: %v", err)
			}
			continue
		default:
		}

		result, err := w.redis.BRPop(ctx, redisPollInterval, w.key).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				continue
			}
			log.Printf("read Redis queue: %v", err)
			if !wait(ctx, redisPollInterval) {
				continue
			}
			continue
		}
		if len(result) != 2 {
			log.Printf("ignore malformed Redis response")
			continue
		}

		var entry logger.LogEntry
		if err := json.Unmarshal([]byte(result[1]), &entry); err != nil {
			log.Printf("discard invalid log entry: %v", err)
			continue
		}
		w.batch = append(w.batch, entry)
	}
}

func (w *Worker) flushWithRetry(ctx context.Context) error {
	if len(w.batch) == 0 {
		return nil
	}
	delay := time.Second
	for {
		if err := w.writer.WriteBatch(ctx, w.batch); err == nil {
			w.batch = w.batch[:0]
			return nil
		} else {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("write batch to Cloud Logging: %v; retrying in %s", err, delay)
		}
		if !wait(ctx, delay) {
			return ctx.Err()
		}
		delay *= 2
		if delay > maxRetryDelay {
			delay = maxRetryDelay
		}
	}
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
