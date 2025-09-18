// Package main wires the RPC client, in-memory storage, parser/poller, and HTTP server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/danieloluwadare/tw-txparser/internal/server"
	"github.com/danieloluwadare/tw-txparser/internal/storage"
	"github.com/danieloluwadare/tw-txparser/pkg/parser"
	"github.com/danieloluwadare/tw-txparser/pkg/rpc"
)

// main is the entry point. It starts the block poller and the HTTP server,
// and performs a graceful shutdown on SIGINT/SIGTERM.
func main() {
	// RPC client - get URL from environment variable with fallback
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://ethereum-rpc.publicnode.com"
	}
	log.Printf("Using Ethereum RPC URL: %s", rpcURL)
	client := rpc.NewClient(rpcURL)

	// In-memory storage
	store := storage.NewMemoryStorage()

	// Config from environment with defaults
	backwardEnabled := true
	if v := os.Getenv("BACKWARD_SCAN_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			backwardEnabled = b
		}
	}
	backwardDepth := 10000
	if v := os.Getenv("BACKWARD_SCAN_DEPTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			backwardDepth = n
		}
	}

	// Parser with options
	p := parser.NewParserWithInterval(client, store, 5*time.Second, parser.Options{
		BackwardScanEnabled: backwardEnabled,
		BackwardScanDepth:   backwardDepth,
	})

	// Cast parserImpl back to Poller
	poller, ok := p.(parser.Poller)
	if !ok {
		log.Fatal("parser does not implement Poller")
	}

	// Create root context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start polling
	log.Println("Starting Poller")
	poller.Start(ctx)

	// Start HTTP API
	s := server.New(p)
	go func() {
		log.Println("Starting server on :8080")
		if err := s.Start(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down...")

	// Cancel context to signal all goroutines to stop
	cancel()

	// Wait for all parser goroutines to complete gracefully
	poller.Stop()
}
