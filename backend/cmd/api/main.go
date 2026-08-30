package main

import (
	"log"

	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/platform/server"
	"go.uber.org/zap"
)

func main() {
	// Bootstrap logger first so startup errors are structured
	zapLog, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer zapLog.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		zapLog.Fatal("failed to load config", zap.Error(err))
	}

	// Switch to development logger for non-production environments
	if cfg.IsDevelopment() {
		devLog, err := zap.NewDevelopment()
		if err != nil {
			zapLog.Fatal("failed to initialize dev logger", zap.Error(err))
		}
		zapLog = devLog
	}

	if err := server.Start(cfg, zapLog); err != nil {
		zapLog.Fatal("server error", zap.Error(err))
	}
}
