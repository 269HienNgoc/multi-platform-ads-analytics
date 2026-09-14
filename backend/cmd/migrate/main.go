// Command migrate applies or rolls back versioned PostgreSQL migrations.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	configFile := flag.String("config", "configs/config.yaml", "path to the YAML configuration file")
	migrationsDirectory := flag.String("migrations", "migrations", "path to the migration directory")
	action := flag.String("action", "up", "migration action: up, down, or version")
	steps := flag.Int("steps", 0, "number of migrations; down requires a positive value")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return app.RunMigration(ctx, *configFile, *migrationsDirectory, *action, *steps)
}
