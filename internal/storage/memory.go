// Package storage contains the in-memory implementation for subscriptions and transactions.
package storage

import (
	"sync"

	"github.com/danieloluwadare/tw-txparser/pkg/models"
)

// MemoryStorage is a thread-safe in-memory implementation of Storage.
type MemoryStorage struct {
	sync.RWMutex
	subs map[string]bool
	txs  map[string][]models.Transaction
}

// NewMemoryStorage creates a fresh MemoryStorage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		subs: make(map[string]bool),
		txs:  make(map[string][]models.Transaction),
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
func (m *MemoryStorage) AddTransaction(addr string, tx models.Transaction) {
	m.Lock()
	defer m.Unlock()
	m.txs[addr] = append(m.txs[addr], tx)
}

// GetTransactions returns the transactions associated with an address.
func (m *MemoryStorage) GetTransactions(addr string) []models.Transaction {
	m.RLock()
	defer m.RUnlock()
	return m.txs[addr]
}

// IsSubscribed checks if an address is registered.
func (m *MemoryStorage) IsSubscribed(addr string) bool {
	m.RLock()
	defer m.RUnlock()
	return m.subs[addr]
}
