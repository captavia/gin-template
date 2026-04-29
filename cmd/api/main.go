package main

import (
	"fmt"
	"log"

	"template/config"
	"template/internal/api/router"
	"template/internal/di"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	injector := di.BuildContainer(cfg)

	r := router.SetupRouter(injector)

	fmt.Printf("Starting API server on :%s\n", cfg.App.Host)
	if err := r.Run(cfg.App.Host); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
