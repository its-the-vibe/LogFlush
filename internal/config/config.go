package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config contains the worker's non-secret settings and runtime credentials.
type Config struct {
	RedisAddress  string        `yaml:"redis_address"`
	RedisUsername string        `yaml:"redis_username"`
	RedisPassword string        `yaml:"-"`
	RedisKey      string        `yaml:"redis_key"`
	GCPProjectID  string        `yaml:"gcp_project_id"`
	GCPLogName    string        `yaml:"gcp_log_name"`
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"-"`
}

type fileConfig struct {
	RedisAddress  string `yaml:"redis_address"`
	RedisUsername string `yaml:"redis_username"`
	RedisKey      string `yaml:"redis_key"`
	GCPProjectID  string `yaml:"gcp_project_id"`
	GCPLogName    string `yaml:"gcp_log_name"`
	BatchSize     int    `yaml:"batch_size"`
	FlushInterval string `yaml:"flush_interval"`
}

// Load reads a YAML configuration file and applies environment overrides.
func Load(path string) (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		RedisAddress:  "localhost:6379",
		RedisKey:      "logflush",
		GCPLogName:    "logflush",
		BatchSize:     100,
		FlushInterval: 2 * time.Second,
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
		var file fileConfig
		if err := yaml.Unmarshal(data, &file); err != nil {
			return Config{}, fmt.Errorf("parse config file: %w", err)
		}
		if file.RedisAddress != "" {
			cfg.RedisAddress = file.RedisAddress
		}
		cfg.RedisUsername = file.RedisUsername
		if file.RedisKey != "" {
			cfg.RedisKey = file.RedisKey
		}
		cfg.GCPProjectID = file.GCPProjectID
		if file.GCPLogName != "" {
			cfg.GCPLogName = file.GCPLogName
		}
		if file.BatchSize != 0 {
			cfg.BatchSize = file.BatchSize
		}
		if file.FlushInterval != "" {
			interval, err := time.ParseDuration(file.FlushInterval)
			if err != nil {
				return Config{}, fmt.Errorf("parse flush_interval: %w", err)
			}
			cfg.FlushInterval = interval
		}
	}

	if value, ok := os.LookupEnv("REDIS_ADDRESS"); ok {
		cfg.RedisAddress = value
	}
	if value, ok := os.LookupEnv("REDIS_USERNAME"); ok {
		cfg.RedisUsername = value
	}
	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")
	if value, ok := os.LookupEnv("REDIS_KEY"); ok {
		cfg.RedisKey = value
	}
	if value, ok := os.LookupEnv("GCP_PROJECT_ID"); ok {
		cfg.GCPProjectID = value
	}
	if value, ok := os.LookupEnv("GCP_LOG_NAME"); ok {
		cfg.GCPLogName = value
	}
	if value, ok := os.LookupEnv("BATCH_SIZE"); ok {
		batchSize, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse BATCH_SIZE: %w", err)
		}
		cfg.BatchSize = batchSize
	}
	if value, ok := os.LookupEnv("FLUSH_INTERVAL"); ok {
		interval, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse flush_interval: %w", err)
		}
		cfg.FlushInterval = interval
	}
	if cfg.RedisAddress == "" || cfg.RedisKey == "" || cfg.GCPProjectID == "" || cfg.GCPLogName == "" {
		return Config{}, errors.New("redis_address, redis_key, gcp_project_id, and gcp_log_name are required")
	}
	if cfg.BatchSize < 1 || cfg.FlushInterval <= 0 {
		return Config{}, errors.New("batch_size and flush_interval must be positive")
	}
	return cfg, nil
}
