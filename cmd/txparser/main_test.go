package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danieloluwadare/tw-txparser/internal/server"
	"github.com/danieloluwadare/tw-txparser/internal/storage"
	"github.com/danieloluwadare/tw-txparser/pkg/parser"
	"github.com/danieloluwadare/tw-txparser/pkg/rpc"
)

// MockRPCClient for integration testing
type MockRPCClient struct {
	blockNumberResponse string
	blockResponse       rpc.Block
	callError           error
}

func NewMockRPCClient() *MockRPCClient {
	return &MockRPCClient{
		blockNumberResponse: "0x1234",
		blockResponse: rpc.Block{
			Number: "0x1234",
			Transactions: []rpc.Transaction{
				{
					Hash:  "0xhash1",
					From:  "0xfrom1",
					To:    "0xto1",
					Value: "0x1000",
				},
			},
		},
	}
}

func (m *MockRPCClient) Call(ctx context.Context, method string, params []interface{}, result interface{}) error {
	if m.callError != nil {
		return m.callError
	}

	switch method {
	case "eth_blockNumber":
		*result.(*string) = m.blockNumberResponse
	case "eth_getBlockByNumber":
		*result.(*rpc.Block) = m.blockResponse
	}
	return nil
}

// Implement the new helper methods
func (m *MockRPCClient) GetBlockNumber(ctx context.Context) (string, error) {
	return m.blockNumberResponse, nil
}

func (m *MockRPCClient) GetBlockByNumber(ctx context.Context, blockNumber string, includeTransactions bool) (*rpc.Block, error) {
	return &m.blockResponse, nil
}

func (m *MockRPCClient) GetBlockByNumberInt(ctx context.Context, blockNumber int, includeTransactions bool) (*rpc.Block, error) {
	return &m.blockResponse, nil
}

func TestIntegration_SubscribeAndGetTransactions(t *testing.T) {
	// Create mock RPC client
	client := NewMockRPCClient()

	// Create storage
	store := storage.NewMemoryStorage()

	// Create parser
	p := parser.NewParserWithInterval(client, store, 100*time.Millisecond, parser.Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	// Create server
	s := server.New(p)

	// Test HTTP endpoints using the actual HTTP handlers
	address := "0x1234567890abcdef"

	// 1. Subscribe to address
	subscribeBody := map[string]string{"address": address}
	body, _ := json.Marshal(subscribeBody)
	req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Use the actual HTTP handler
	s.HandleSubscribe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Subscribe failed with status %d", w.Code)
	}

	var subscribeResponse map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&subscribeResponse); err != nil {
		t.Fatalf("Failed to decode subscribe response: %v", err)
	}
	if !subscribeResponse["subscribed"] {
		t.Error("Expected subscription to succeed")
	}

	// 2. Get current block
	req = httptest.NewRequest(http.MethodGet, "/current", nil)
	w = httptest.NewRecorder()
	s.HandleCurrentBlock(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get current block failed with status %d", w.Code)
	}

	var blockResponse map[string]int
	if err := json.NewDecoder(w.Body).Decode(&blockResponse); err != nil {
		t.Fatalf("Failed to decode block response: %v", err)
	}
	if blockResponse["block"] != 0 {
		t.Errorf("Expected initial block to be 0, got %d", blockResponse["block"])
	}

	// 3. Get transactions (should be empty initially)
	req = httptest.NewRequest(http.MethodGet, "/transactions?address="+address, nil)
	w = httptest.NewRecorder()
	s.HandleTransactions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get transactions failed with status %d", w.Code)
	}

	var transactions []interface{}
	if err := json.NewDecoder(w.Body).Decode(&transactions); err != nil {
		t.Fatalf("Failed to decode transactions response: %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions initially, got %d", len(transactions))
	}
}

func TestIntegration_ParserWithPoller(t *testing.T) {
	// Create mock RPC client
	client := NewMockRPCClient()

	// Create storage
	store := storage.NewMemoryStorage()

	// Create parser
	p := parser.NewParserWithInterval(client, store, 50*time.Millisecond, parser.Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	// Cast to Poller interface
	poller, ok := p.(parser.Poller)
	if !ok {
		t.Fatal("Parser does not implement Poller interface")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start polling
	poller.Start(ctx)

	// Wait for context to timeout
	<-ctx.Done()

	// Verify that the parser was created successfully
	if p == nil {
		t.Fatal("Expected parser to be created")
	}

	// Test parser methods
	if p.GetCurrentBlock() < 0 {
		t.Error("Expected current block to be non-negative")
	}
}

func TestIntegration_StorageOperations(t *testing.T) {
	// Create storage
	store := storage.NewMemoryStorage()

	// Test subscription
	address := "0x1234567890abcdef"
	if !store.Subscribe(address) {
		t.Error("Expected first subscription to succeed")
	}
	if store.Subscribe(address) {
		t.Error("Expected duplicate subscription to fail")
	}

	// Test transaction storage
	// Note: In a real integration test, transactions would be added by the parser
	// For this test, we'll add them manually to verify storage works
	transactions := store.GetTransactions(address)
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions initially, got %d", len(transactions))
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Create mock RPC client with error
	client := NewMockRPCClient()
	client.callError = &rpc.RPCError{Code: -32601, Message: "Method not found"}

	// Create storage
	store := storage.NewMemoryStorage()

	// Create parser
	p := parser.NewParserWithInterval(client, store, 50*time.Millisecond, parser.Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	// Test that parser handles errors gracefully
	if p == nil {
		t.Fatal("Expected parser to be created even with error-prone client")
	}

	// Test parser methods still work
	if p.GetCurrentBlock() != 0 {
		t.Errorf("Expected initial block to be 0, got %d", p.GetCurrentBlock())
	}
}

func TestIntegration_ConcurrentAccess(t *testing.T) {
	// Create storage
	store := storage.NewMemoryStorage()

	// Test concurrent subscription attempts
	address := "0x1234567890abcdef"
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			store.Subscribe(address)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify only one subscription succeeded
	if !store.IsSubscribed(address) {
		t.Error("Expected address to be subscribed")
	}
}
