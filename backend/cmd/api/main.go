// Command api starts the multi-platform advertising analytics HTTP API.
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
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return app.Run(ctx, *configFile)
}
