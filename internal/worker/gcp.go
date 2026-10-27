package worker

import (
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/logging"
	"github.com/its-the-vibe/LogFlush/pkg/logger"
	"google.golang.org/api/option"
)

type logWriter interface {
	WriteBatch(context.Context, []logger.LogEntry) error
}

type cloudWriter struct {
	client *logging.Client
	logger *logging.Logger
}

// NewCloudWriter connects to Google Cloud Logging. If credentialsFile is empty,
// application default credentials are used.
func NewCloudWriter(ctx context.Context, projectID, logName, credentialsFile string) (*cloudWriter, error) {
	if projectID == "" || logName == "" {
		return nil, errors.New("GCP project ID and log name are required")
	}
	var opts []option.ClientOption
	if credentialsFile != "" {
		info, err := os.Stat(credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("GCP credentials file: %w", err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("GCP credentials file %q is a directory", credentialsFile)
		}
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	client, err := logging.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}
	return &cloudWriter{client: client, logger: client.Logger(logName)}, nil
}

// Close releases the underlying Cloud Logging client.
func (w *cloudWriter) Close() error {
	return w.client.Close()
}

func (w *cloudWriter) WriteBatch(ctx context.Context, entries []logger.LogEntry) error {
	for _, entry := range entries {
		w.logger.Log(toCloudEntry(entry))
	}
	return w.logger.Flush()
}

func toCloudEntry(entry logger.LogEntry) logging.Entry {
	return logging.Entry{
		Timestamp: entry.Timestamp,
		Severity:  cloudSeverity(entry.Severity),
		Payload: map[string]any{
			"message":    entry.Message,
			"service":    entry.Service,
			"attributes": entry.Attributes,
		},
	}
}

func cloudSeverity(severity string) logging.Severity {
	switch severity {
	case "DEBUG":
		return logging.Debug
	case "INFO":
		return logging.Info
	case "WARNING", "WARN":
		return logging.Warning
	case "ERROR":
		return logging.Error
	case "CRITICAL", "ALERT", "EMERGENCY":
		return logging.Critical
	default:
		return logging.Default
	}
}
