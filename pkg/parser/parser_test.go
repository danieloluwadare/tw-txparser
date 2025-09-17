package parser

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/danieloluwadare/tw-txparser/pkg/rpc"
	"github.com/danieloluwadare/tw-txparser/pkg/transaction"
)

// MockStorage implements the storage.Storage interface for testing
type MockStorage struct {
	subscriptions map[string]bool
	transactions  map[string][]transaction.Transaction
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		subscriptions: make(map[string]bool),
		transactions:  make(map[string][]transaction.Transaction),
	}
}

func (m *MockStorage) Subscribe(address string) bool {
	if m.subscriptions[address] {
		return false
	}
	m.subscriptions[address] = true
	return true
}

func (m *MockStorage) AddTransaction(addr string, tx transaction.Transaction) {
	m.transactions[addr] = append(m.transactions[addr], tx)
}

func (m *MockStorage) GetTransactions(address string) []transaction.Transaction {
	return m.transactions[address]
}

func (m *MockStorage) IsSubscribed(addr string) bool {
	return m.subscriptions[addr]
}

// MockRPCClient implements a mock RPC client for testing
type MockRPCClient struct {
	blockNumberResponse string
	blockResponse       rpc.Block
	callError           error
	callCount           int
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
				{
					Hash:  "0xhash2",
					From:  "0xfrom2",
					To:    "0xto2",
					Value: "0x2000",
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
		m.callCount++
		// Return increasing block numbers for first few calls, then stable
		if m.callCount <= 3 {
			blockNum := 0x1234 + m.callCount
			*result.(*string) = fmt.Sprintf("0x%x", blockNum)
		} else {
			// Return stable block number to prevent infinite processing
			*result.(*string) = "0x1237"
		}
	case "eth_getBlockByNumber":
		*result.(*rpc.Block) = m.blockResponse
	}
	return nil
}

func TestNewParserWithInterval(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	interval := 5 * time.Second

	parser := NewParserWithInterval(client, store, interval, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	// Verify parser implements Parser interface
	if parser == nil {
		t.Fatal("Expected parser to be created")
	}

	// Test initial state
	if parser.GetCurrentBlock() != 0 {
		t.Errorf("Expected initial block to be 0, got %d", parser.GetCurrentBlock())
	}
}

func TestParser_GetCurrentBlock(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 5*time.Second, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	// Test initial block
	block := parser.GetCurrentBlock()
	if block != 0 {
		t.Errorf("Expected initial block to be 0, got %d", block)
	}
}

func TestParser_Subscribe(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 5*time.Second, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	address := "0x1234567890abcdef"

	// Test subscribing to new address
	result := parser.Subscribe(address)
	if !result {
		t.Error("Expected Subscribe to return true for new address")
	}

	// Test subscribing to same address again
	result = parser.Subscribe(address)
	if result {
		t.Error("Expected Subscribe to return false for already subscribed address")
	}
}

func TestParser_GetTransactions(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 5*time.Second, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	address := "0x1234567890abcdef"

	// Test getting transactions for non-existent address
	transactions := parser.GetTransactions(address)
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions for new address, got %d", len(transactions))
	}

	// Add some transactions directly to storage
	tx1 := transaction.Transaction{Hash: "0xhash1", From: "0xfrom1", To: address, Value: "1000", Block: 1, Inbound: true}
	tx2 := transaction.Transaction{Hash: "0xhash2", From: "0xfrom2", To: address, Value: "2000", Block: 2, Inbound: true}

	store.AddTransaction(address, tx1)
	store.AddTransaction(address, tx2)

	// Test getting transactions
	transactions = parser.GetTransactions(address)
	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}

	if transactions[0].Hash != tx1.Hash {
		t.Errorf("Expected first transaction hash %s, got %s", tx1.Hash, transactions[0].Hash)
	}
	if transactions[1].Hash != tx2.Hash {
		t.Errorf("Expected second transaction hash %s, got %s", tx2.Hash, transactions[1].Hash)
	}
}

func TestParser_Start(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 100*time.Millisecond, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	// Cast to parserImpl to access Start method
	parserImpl, ok := parser.(*parserImpl)
	if !ok {
		t.Fatal("Expected parser to be of type *parserImpl")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start the parser
	parserImpl.Start(ctx)

	// Wait for context to timeout
	<-ctx.Done()

	// Verify that polling was started
	parserImpl.pollingStartedMu.Lock()
	started := parserImpl.pollingStarted
	parserImpl.pollingStartedMu.Unlock()

	if !started {
		t.Error("Expected polling to be started")
	}
}

func TestParser_Start_MultipleCalls(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 50*time.Millisecond, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	parserImpl, ok := parser.(*parserImpl)
	if !ok {
		t.Fatal("Expected parser to be of type *parserImpl")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	// Start the parser multiple times
	parserImpl.Start(ctx)
	parserImpl.Start(ctx)
	parserImpl.Start(ctx)

	// Wait for context to timeout
	<-ctx.Done()

	// Verify that polling was started only once
	parserImpl.pollingStartedMu.Lock()
	started := parserImpl.pollingStarted
	parserImpl.pollingStartedMu.Unlock()

	if !started {
		t.Error("Expected polling to be started")
	}
}

func TestParser_Stop(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 50*time.Millisecond, Options{BackwardScanEnabled: false, BackwardScanDepth: 10000})

	parserImpl, ok := parser.(*parserImpl)
	if !ok {
		t.Fatal("Expected parser to be of type *parserImpl")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start the parser
	parserImpl.Start(ctx)

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Cancel the context to signal shutdown
	cancel()

	// Stop the parser - this should block until all goroutines complete
	start := time.Now()
	parserImpl.Stop()
	duration := time.Since(start)

	// Verify that polling was stopped
	parserImpl.pollingStartedMu.Lock()
	started := parserImpl.pollingStarted
	parserImpl.pollingStartedMu.Unlock()

	if started {
		t.Error("Expected polling to be stopped")
	}

	// The stop should have completed quickly (not hanging)
	if duration > 100*time.Millisecond {
		t.Errorf("Stop took too long: %v", duration)
	}
}

func TestFormatBlockNum(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0x0"},
		{1, "0x1"},
		{10, "0xa"},
		{16, "0x10"},
		{255, "0xff"},
		{256, "0x100"},
		{4095, "0xfff"},
		{4096, "0x1000"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := formatBlockNum(tt.input)
			if result != tt.expected {
				t.Errorf("formatBlockNum(%d) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessBlock(t *testing.T) {
	client := NewMockRPCClient()
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 5*time.Second, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	parserImpl, ok := parser.(*parserImpl)
	if !ok {
		t.Fatal("Expected parser to be of type *parserImpl")
	}

	// Process a block - all transactions are stored regardless of subscription status
	err := parserImpl.processBlock(context.Background(), 1234)
	if err != nil {
		t.Fatalf("processBlock failed: %v", err)
	}

	// Verify transactions were added to storage
	// All transactions are stored regardless of subscription status
	from1Txs := store.GetTransactions("0xfrom1")
	to1Txs := store.GetTransactions("0xto1")
	from2Txs := store.GetTransactions("0xfrom2")
	to2Txs := store.GetTransactions("0xto2")

	if len(from1Txs) != 1 {
		t.Errorf("Expected 1 transaction for from1, got %d", len(from1Txs))
	}
	if len(to1Txs) != 1 {
		t.Errorf("Expected 1 transaction for to1, got %d", len(to1Txs))
	}
	if len(from2Txs) != 1 {
		t.Errorf("Expected 1 transaction for from2, got %d", len(from2Txs))
	}
	if len(to2Txs) != 1 {
		t.Errorf("Expected 1 transaction for to2, got %d", len(to2Txs))
	}

	// Verify transaction details for from1 (outbound transaction)
	tx := from1Txs[0]
	if tx.Hash != "0xhash1" {
		t.Errorf("Expected hash 0xhash1, got %s", tx.Hash)
	}
	if tx.Block != 1234 {
		t.Errorf("Expected block 1234, got %d", tx.Block)
	}
	if tx.Value != "4096" { // 0x1000 in decimal
		t.Errorf("Expected value 4096, got %s", tx.Value)
	}
	if tx.Inbound != false {
		t.Errorf("Expected Inbound=false for from1 transaction, got %t", tx.Inbound)
	}

	// Verify transaction details for to1 (inbound transaction)
	tx = to1Txs[0]
	if tx.Hash != "0xhash1" {
		t.Errorf("Expected hash 0xhash1, got %s", tx.Hash)
	}
	if tx.Block != 1234 {
		t.Errorf("Expected block 1234, got %d", tx.Block)
	}
	if tx.Value != "4096" { // 0x1000 in decimal
		t.Errorf("Expected value 4096, got %s", tx.Value)
	}
	if tx.Inbound != true {
		t.Errorf("Expected Inbound=true for to1 transaction, got %t", tx.Inbound)
	}
}

func TestProcessBlock_Error(t *testing.T) {
	client := NewMockRPCClient()
	client.callError = &rpc.RPCError{Code: -32601, Message: "Method not found"}
	store := NewMockStorage()
	parser := NewParserWithInterval(client, store, 5*time.Second, Options{BackwardScanEnabled: true, BackwardScanDepth: 10000})

	parserImpl, ok := parser.(*parserImpl)
	if !ok {
		t.Fatal("Expected parser to be of type *parserImpl")
	}

	// Process a block with error
	err := parserImpl.processBlock(context.Background(), 1234)
	if err == nil {
		t.Error("Expected processBlock to return error")
	}

	// Verify no transactions were added
	from1Txs := store.GetTransactions("0xfrom1")
	if len(from1Txs) != 0 {
		t.Errorf("Expected 0 transactions for from1 due to error, got %d", len(from1Txs))
	}
}
