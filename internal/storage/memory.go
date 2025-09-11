// Package storage contains the in-memory implementation for subscriptions and transactions.
package storage

import (
	"github.com/danieloluwadare/tw-txparser/pkg/transaction"
	"sync"
)

// MemoryStorage is a thread-safe in-memory implementation of Storage.
type MemoryStorage struct {
	sync.RWMutex
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
	m.Lock()
	defer m.Unlock()
	if m.subs[address] {
		return false
	}
	m.subs[address] = true
	return true
}

// AddTransaction appends a transaction to an address's list.
func (m *MemoryStorage) AddTransaction(addr string, tx transaction.Transaction) {
	m.Lock()
	defer m.Unlock()
	m.txs[addr] = append(m.txs[addr], tx)
}

// GetTransactions returns the transactions associated with an address.
// Only returns transactions if the address is subscribed.
func (m *MemoryStorage) GetTransactions(addr string) []transaction.Transaction {
	m.RLock()
	defer m.RUnlock()

	// Only return transactions if address is subscribed
	if !m.subs[addr] {
		return []transaction.Transaction{}
	}
	return m.txs[addr]
}

// IsSubscribed checks if an address is registered.
func (m *MemoryStorage) IsSubscribed(addr string) bool {
	m.RLock()
	defer m.RUnlock()
	return m.subs[addr]
}
