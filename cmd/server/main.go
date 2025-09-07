package main

import (
	"log"
	"os"

	"github.com/fleshka4/1inch-test-task/internal/config"
	"github.com/fleshka4/1inch-test-task/internal/infra/uniswap"
	"github.com/fleshka4/1inch-test-task/internal/service"
	"github.com/fleshka4/1inch-test-task/internal/transport/http"
)

func main() {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "cfg/config.yaml"
	}

	cfg := config.Load(path)

	client, err := uniswap.NewClient(cfg.RPCURL)
	if err != nil {
		log.Fatalf("uniswap.NewClient: %v", err)
	}

	estimator := service.NewEstimatorService(client)

	srv := http.NewServer(estimator, cfg)

	err = srv.ListenAndServe(cfg.ListenAddr)
	if err != nil {
		log.Fatalf("srv.ListenAndServe: %v", err)
	}
}
