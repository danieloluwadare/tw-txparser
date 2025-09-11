// Package storage defines the storage interfaces.
package storage

import "github.com/danieloluwadare/tw-txparser/pkg/transaction"

// Storage abstracts subscriptions and per-address transactions.
type Storage interface {
	// Subscribe registers an address and returns false if it already existed.
	Subscribe(address string) bool
	// AddTransaction appends a transaction for the given address.
	AddTransaction(addr string, tx transaction.Transaction)
	// GetTransactions returns transactions associated with address.
	GetTransactions(address string) []transaction.Transaction
	// IsSubscribed indicates whether address is registered.
	IsSubscribed(addr string) bool
}
