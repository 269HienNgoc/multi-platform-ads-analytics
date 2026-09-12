package main

import (
	"log"
	"net/http"
	"os"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/httpapi"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/workflow"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	service := workflow.NewService()
	server := httpapi.NewServer(service)

	log.Printf("campaign automation API listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
