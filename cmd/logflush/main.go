package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/its-the-vibe/LogFlush/internal/config"
	"github.com/its-the-vibe/LogFlush/internal/worker"
	"github.com/redis/go-redis/v9"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("connect to Redis: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cloudWriter, err := worker.NewCloudWriter(context.Background(), cfg.GCPProjectID, cfg.GCPLogName)
	if err != nil {
		log.Fatalf("create Cloud Logging client: %v", err)
	}
	defer cloudWriter.Close()

	logWorker, err := worker.New(redisClient, cfg.RedisKey, cloudWriter, cfg.BatchSize, cfg.FlushInterval)
	if err != nil {
		log.Fatalf("create worker: %v", err)
	}
	if err := logWorker.Run(ctx); err != nil {
		log.Fatalf("worker stopped: %v", err)
	}
}
