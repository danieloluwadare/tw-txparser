package storage

import (
	"testing"

	"github.com/danieloluwadare/tw-txparser/pkg/models"
)

func TestMemoryStorage_Subscribe(t *testing.T) {
	store := NewMemoryStorage()

	// Test subscribing to a new address
	address := "0x1234567890abcdef"
	result := store.Subscribe(address)
	if !result {
		t.Error("Expected Subscribe to return true for new address")
	}

	// Verify the address is subscribed
	if !store.IsSubscribed(address) {
		t.Error("Expected address to be subscribed")
	}

	// Test subscribing to the same address again
	result = store.Subscribe(address)
	if result {
		t.Error("Expected Subscribe to return false for already subscribed address")
	}

	// Test subscribing to a different address
	address2 := "0xfedcba0987654321"
	result = store.Subscribe(address2)
	if !result {
		t.Error("Expected Subscribe to return true for new address")
	}

	// Verify both addresses are subscribed
	if !store.IsSubscribed(address) {
		t.Error("Expected first address to still be subscribed")
	}
	if !store.IsSubscribed(address2) {
		t.Error("Expected second address to be subscribed")
	}
}

func TestMemoryStorage_AddTransaction(t *testing.T) {
	store := NewMemoryStorage()
	address := "0x1234567890abcdef"

	// Subscribe to address first
	store.Subscribe(address)

	// Add first transaction
	tx1 := models.Transaction{
		Hash:    "0xhash1",
		From:    "0xfrom1",
		To:      address,
		Value:   "1000",
		Block:   1,
		Inbound: true,
	}
	store.AddTransaction(address, tx1)

	// Verify transaction was added
	transactions := store.GetTransactions(address)
	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}
	if transactions[0].Hash != tx1.Hash {
		t.Errorf("Expected hash %s, got %s", tx1.Hash, transactions[0].Hash)
	}

	// Add second transaction
	tx2 := models.Transaction{
		Hash:    "0xhash2",
		From:    "0xfrom2",
		To:      address,
		Value:   "2000",
		Block:   2,
		Inbound: true,
	}
	store.AddTransaction(address, tx2)

	// Verify both transactions are present
	transactions = store.GetTransactions(address)
	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}

	// Verify order (should be in order added)
	if transactions[0].Hash != tx1.Hash {
		t.Errorf("Expected first transaction hash %s, got %s", tx1.Hash, transactions[0].Hash)
	}
	if transactions[1].Hash != tx2.Hash {
		t.Errorf("Expected second transaction hash %s, got %s", tx2.Hash, transactions[1].Hash)
	}
}

func TestMemoryStorage_GetTransactions(t *testing.T) {
	store := NewMemoryStorage()
	address := "0x1234567890abcdef"

	// Test getting transactions for non-existent address
	transactions := store.GetTransactions(address)
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions for new address, got %d", len(transactions))
	}

	// Subscribe to address first
	store.Subscribe(address)

	// Add some transactions
	tx1 := models.Transaction{Hash: "0xhash1", From: "0xfrom1", To: address, Value: "1000", Block: 1, Inbound: true}
	tx2 := models.Transaction{Hash: "0xhash2", From: "0xfrom2", To: address, Value: "2000", Block: 2, Inbound: true}

	store.AddTransaction(address, tx1)
	store.AddTransaction(address, tx2)

	// Test getting transactions
	transactions = store.GetTransactions(address)
	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}

	// Verify transaction details
	if transactions[0].Hash != tx1.Hash {
		t.Errorf("Expected first transaction hash %s, got %s", tx1.Hash, transactions[0].Hash)
	}
	if transactions[1].Hash != tx2.Hash {
		t.Errorf("Expected second transaction hash %s, got %s", tx2.Hash, transactions[1].Hash)
	}
}

func TestMemoryStorage_GetTransactions_SubscriptionRequired(t *testing.T) {
	store := NewMemoryStorage()
	address := "0x1234567890abcdef"

	// Add transactions without subscribing
	tx1 := models.Transaction{Hash: "0xhash1", From: "0xfrom1", To: address, Value: "1000", Block: 1, Inbound: true}
	tx2 := models.Transaction{Hash: "0xhash2", From: "0xfrom2", To: address, Value: "2000", Block: 2, Inbound: true}

	store.AddTransaction(address, tx1)
	store.AddTransaction(address, tx2)

	// GetTransactions should return empty for unsubscribed address
	transactions := store.GetTransactions(address)
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions for unsubscribed address, got %d", len(transactions))
	}

	// Subscribe to address
	store.Subscribe(address)

	// Now GetTransactions should return the transactions
	transactions = store.GetTransactions(address)
	if len(transactions) != 2 {
		t.Errorf("Expected 2 transactions for subscribed address, got %d", len(transactions))
	}
}

func TestMemoryStorage_IsSubscribed(t *testing.T) {
	store := NewMemoryStorage()
	address := "0x1234567890abcdef"

	// Test non-subscribed address
	if store.IsSubscribed(address) {
		t.Error("Expected address to not be subscribed initially")
	}

	// Subscribe to address
	store.Subscribe(address)

	// Test subscribed address
	if !store.IsSubscribed(address) {
		t.Error("Expected address to be subscribed after Subscribe call")
	}

	// Test different address
	address2 := "0xfedcba0987654321"
	if store.IsSubscribed(address2) {
		t.Error("Expected different address to not be subscribed")
	}
}

func TestMemoryStorage_Concurrency(t *testing.T) {
	store := NewMemoryStorage()
	address := "0x1234567890abcdef"

	// Subscribe to address first
	store.Subscribe(address)

	// Test concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines that add transactions
	for i := 0; i < 5; i++ {
		go func(i int) {
			tx := models.Transaction{
				Hash:    "0xhash" + string(rune(i)),
				From:    "0xfrom",
				To:      address,
				Value:   "1000",
				Block:   i,
				Inbound: true,
			}
			store.AddTransaction(address, tx)
			done <- true
		}(i)
	}

	// Start multiple goroutines that read transactions
	for i := 0; i < 5; i++ {
		go func() {
			store.GetTransactions(address)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	transactions := store.GetTransactions(address)
	if len(transactions) != 5 {
		t.Errorf("Expected 5 transactions after concurrent access, got %d", len(transactions))
	}
}

func TestMemoryStorage_MultipleAddresses(t *testing.T) {
	store := NewMemoryStorage()
	address1 := "0x1234567890abcdef"
	address2 := "0xfedcba0987654321"

	// Subscribe to both addresses
	store.Subscribe(address1)
	store.Subscribe(address2)

	// Add transactions for different addresses
	tx1 := models.Transaction{Hash: "0xhash1", From: "0xfrom1", To: address1, Value: "1000", Block: 1, Inbound: true}
	tx2 := models.Transaction{Hash: "0xhash2", From: "0xfrom2", To: address2, Value: "2000", Block: 2, Inbound: true}

	store.AddTransaction(address1, tx1)
	store.AddTransaction(address2, tx2)

	// Verify transactions are stored separately
	transactions1 := store.GetTransactions(address1)
	transactions2 := store.GetTransactions(address2)

	if len(transactions1) != 1 {
		t.Errorf("Expected 1 transaction for address1, got %d", len(transactions1))
	}
	if len(transactions2) != 1 {
		t.Errorf("Expected 1 transaction for address2, got %d", len(transactions2))
	}

	if transactions1[0].Hash != tx1.Hash {
		t.Errorf("Expected transaction1 hash %s, got %s", tx1.Hash, transactions1[0].Hash)
	}
	if transactions2[0].Hash != tx2.Hash {
		t.Errorf("Expected transaction2 hash %s, got %s", tx2.Hash, transactions2[0].Hash)
	}
}
