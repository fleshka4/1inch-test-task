package main

import (
	"log"
	"os"

	"github.com/fleshka4/inch-test-task/internal/httpserver"
)

func main() {
	rpc := os.Getenv("ETH_RPC_URL")
	if rpc == "" {
		log.Fatal("ETH_RPC_URL не задан")
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":1337"
	}

	srv, err := httpserver.New(rpc)
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	log.Printf("listening on %s", addr)
	log.Fatal(srv.ListenAndServe(addr))
}
