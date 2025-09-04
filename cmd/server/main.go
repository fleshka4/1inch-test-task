package main

import (
	"log"
	"os"

	"github.com/fleshka4/inch-test-task/internal/config"
	"github.com/fleshka4/inch-test-task/internal/httpserver"
	"github.com/fleshka4/inch-test-task/internal/uniswapv2"
)

func main() {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "cfg/config.yaml"
	}

	cfg := config.Load(path)

	client, err := uniswapv2.NewClient(cfg.RPCURL)
	if err != nil {
		log.Fatalf("uniswapv2.NewClient: %v", err)
	}

	srv, err := httpserver.New(client, cfg)
	if err != nil {
		log.Fatalf("httpserver.New: %v", err)
	}

	err = srv.ListenAndServe(cfg.ListenAddr)
	if err != nil {
		log.Fatalf("srv.ListenAndServe: %v", err)
	}
}
