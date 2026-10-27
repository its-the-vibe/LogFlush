# LogFlush

[![CI](https://github.com/its-the-vibe/LogFlush/actions/workflows/ci.yaml/badge.svg)](https://github.com/its-the-vibe/LogFlush/actions/workflows/ci.yaml)

LogFlush is a lightweight logging proxy. Applications push structured log entries to an externally hosted Redis list; the worker reads entries in batches and forwards them to Google Cloud Logging.

## Architecture

- `pkg/logger` provides the Go client and `LogEntry` type used by applications.
- `cmd/logflush` starts the worker.
- `internal/config` loads YAML settings and environment overrides.
- `internal/worker` consumes Redis with `BRPOP`, batches entries, and retries Cloud Logging writes with exponential backoff.

The client uses `LPUSH` and the worker uses `BRPOP`, preserving FIFO order. Delivery is at-least-once when Cloud Logging write results are ambiguous: a failed batch is retried and may be duplicated if the remote service accepted it before the failure was observed.

## Configuration and credentials

Copy the examples before running locally:

```sh
cp config.example.yaml config.yaml
cp .env.example .env
```

Set `gcp_project_id` in `config.yaml` and update `redis_address`, `redis_key`, `gcp_log_name`, `batch_size`, and `flush_interval` as needed. Do not put credentials in the YAML file. Put `REDIS_PASSWORD` in `.env`; `.env` and `config.yaml` are gitignored. Environment variables override YAML values. Supported overrides include `REDIS_ADDRESS`, `REDIS_USERNAME`, `REDIS_PASSWORD`, `REDIS_KEY`, `GCP_PROJECT_ID`, `GCP_LOG_NAME`, `GCP_CREDENTIALS_FILE`, `BATCH_SIZE`, and `FLUSH_INTERVAL`.

Optionally set `gcp_credentials_file` in `config.yaml` (or `GCP_CREDENTIALS_FILE`) to the path of a service account JSON key file, for example `gcp_credentials_file: /secrets/gcp-key.json`. If the file is missing or invalid, the worker fails at startup. Mount the key read-only and never commit it.

When no credentials file is set, the worker uses Google Application Default Credentials. For local development, authenticate with `gcloud auth application-default login`; in production, use the runtime's attached service account with permission to write Cloud Logging entries. Redis must be reachable from the worker.

## Client example

```go
redisClient := redis.NewClient(&redis.Options{
    Addr:     "redis.example.com:6379",
    Password: os.Getenv("REDIS_PASSWORD"),
})
defer redisClient.Close()

logClient, err := logger.New(redisClient, "logflush")
if err != nil {
    return err
}
err = logClient.Log(ctx, logger.LogEntry{
    Severity:  "INFO",
    Message:   "payment accepted",
    Service:   "payments",
    Timestamp: time.Now().UTC(),
    Attributes: map[string]any{
        "payment_id": "p-123",
    },
})
```

Import `github.com/its-the-vibe/LogFlush/pkg/logger` and `github.com/redis/go-redis/v9` in the application using the client.

## Local development

Install Go 1.27 or later, configure `config.yaml` and `.env`, then run:

```sh
make test
make lint
make build
make run
```

The worker handles SIGINT and SIGTERM by stopping queue consumption and flushing buffered entries.

## Docker Compose deployment

Configure `config.yaml` and `.env` as above, then build and start the worker:

```sh
docker compose up --build
```

Compose runs LogFlush with a read-only root filesystem and mounts the configuration read-only. Redis is external and is not started by this Compose file. The container uses a distroless non-root runtime; provide Google credentials through the deployment platform's workload identity or mounted application-default credentials.
