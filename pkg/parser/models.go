// Package parser contains the block poller and parsing logic.
package parser

import (
	"context"

	"github.com/danieloluwadare/tw-txparser/pkg/models"
)

// Parser exposes read APIs and subscription management.
type Parser interface {
	// GetCurrentBlock returns the last processed block number.
	GetCurrentBlock() int
	// Subscribe registers an address to track.
	Subscribe(address string) bool
	// GetTransactions lists transactions associated with the address.
	GetTransactions(address string) []models.Transaction
}

// Poller drives continuous block polling until the context is cancelled.
type Poller interface {
	Start(ctx context.Context)
}
