// Package storage contains the in-memory implementation for subscriptions and transactions.
package storage

import (
	"sync"

	"github.com/danieloluwadare/tw-txparser/pkg/transaction"
)

// MemoryStorage is a thread-safe in-memory implementation of Storage.
type MemoryStorage struct {
	mu   sync.Mutex
	subs map[string]bool
	txs  map[string][]transaction.Transaction
}

// NewMemoryStorage creates a fresh MemoryStorage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		subs: make(map[string]bool),
		txs:  make(map[string][]transaction.Transaction),
	}
}

// Subscribe registers an address. Returns false if already subscribed.
func (m *MemoryStorage) Subscribe(address string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.subs[address] {
		return false
	}
	m.subs[address] = true
	return true
}

// AddTransaction appends a transaction to an address's list.
func (m *MemoryStorage) AddTransaction(addr string, tx transaction.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs[addr] = append(m.txs[addr], tx)
}

// GetTransactions returns the transactions associated with an address.
// Only returns transactions if the address is subscribed.
func (m *MemoryStorage) GetTransactions(addr string) []transaction.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only return transactions if address is subscribed
	if !m.subs[addr] {
		return []transaction.Transaction{}
	}
	return m.txs[addr]
}

// IsSubscribed checks if an address is registered.
func (m *MemoryStorage) IsSubscribed(addr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.subs[addr]
}
