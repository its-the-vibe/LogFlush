package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadReadsConfigAndEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("redis_address: redis.example:6379\nredis_key: app-logs\ngcp_project_id: project-from-file\ngcp_log_name: app\nbatch_size: 25\nflush_interval: 500ms\n")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GCP_PROJECT_ID", "project-from-env")
	t.Setenv("REDIS_PASSWORD", "test-password")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisAddress != "redis.example:6379" || cfg.RedisKey != "app-logs" ||
		cfg.GCPProjectID != "project-from-env" || cfg.RedisPassword != "test-password" ||
		cfg.BatchSize != 25 || cfg.FlushInterval != 500*time.Millisecond {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadRejectsInvalidFlushInterval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("gcp_project_id: project\nflush_interval: not-a-duration\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid flush interval error")
	}
}
