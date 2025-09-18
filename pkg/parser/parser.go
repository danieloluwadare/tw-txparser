// Package parser contains the block poller and parsing logic.
package parser

import (
	"sync"
	"time"

	"github.com/danieloluwadare/tw-txparser/internal/storage"
	"github.com/danieloluwadare/tw-txparser/pkg/rpc"
	"github.com/danieloluwadare/tw-txparser/pkg/transaction"
)

// parserImpl implements Parser and Poller using an RPC client and Storage.
type parserImpl struct {
	client           rpc.RPCClient
	store            storage.Storage
	block            int
	pollingStarted   bool
	pollingStartedMu sync.Mutex
	pollInterval     time.Duration
	// goroutine management
	wg sync.WaitGroup
	// configuration
	backwardScanEnabled bool
	backwardScanDepth   int
}

// Options configures parserImpl behavior.
type Options struct {
	BackwardScanEnabled bool
	BackwardScanDepth   int
}

// NewParserWithInterval constructs a parser with a polling interval.
func NewParserWithInterval(c rpc.RPCClient, s storage.Storage, interval time.Duration, opts Options) Parser {
	// apply defaults
	if opts.BackwardScanDepth <= 0 {
		opts.BackwardScanDepth = 10000
	}
	// default enabled = true unless explicitly set false
	// zero value for bool is false; we want default true. Detect "unset" via separate flag? Keep simple: default true if depth>0 and not explicitly false.
	enabled := true
	if !opts.BackwardScanEnabled {
		enabled = false
	}

	return &parserImpl{
		client:              c,
		store:               s,
		block:               0,
		pollInterval:        interval,
		backwardScanEnabled: enabled,
		backwardScanDepth:   opts.BackwardScanDepth,
	}
}

// GetCurrentBlock returns the last processed block number.
func (p *parserImpl) GetCurrentBlock() int {
	return p.block
}

// Subscribe registers an address with the underlying storage.
func (p *parserImpl) Subscribe(address string) bool {
	return p.store.Subscribe(address)
}

// GetTransactions returns transactions from the underlying storage.
func (p *parserImpl) GetTransactions(address string) []transaction.Transaction {
	return p.store.GetTransactions(address)
}
