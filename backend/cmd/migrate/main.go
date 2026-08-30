package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/harmanto-49/cankora/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("migrate: config load: %v", err)
	}

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	// golang-migrate expects postgres:// DSN format
	dsn := "postgres://" + cfg.Database.User + ":" + cfg.Database.Password +
		"@" + cfg.Database.Host + ":" + cfg.Database.Port +
		"/" + cfg.Database.Name + "?sslmode=" + cfg.Database.SSLMode

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatalf("migrate: init: %v", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrate: all up migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("migrate: all down migrations applied")
	case "drop":
		if err := m.Drop(); err != nil {
			log.Fatalf("migrate drop: %v", err)
		}
		log.Println("migrate: database dropped")
	default:
		log.Fatalf("migrate: unknown direction %q (use: up | down | drop)", direction)
	}
}
