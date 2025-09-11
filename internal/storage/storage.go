// Package storage defines the storage interfaces.
package storage

import "github.com/danieloluwadare/tw-txparser/pkg/models"

// Storage abstracts subscriptions and per-address transactions.
type Storage interface {
	// Subscribe registers an address and returns false if it already existed.
	Subscribe(address string) bool
	// AddTransaction appends a transaction for the given address.
	AddTransaction(addr string, tx models.Transaction)
	// GetTransactions returns transactions associated with address.
	GetTransactions(address string) []models.Transaction
	// IsSubscribed indicates whether address is registered.
	IsSubscribed(addr string) bool
}
