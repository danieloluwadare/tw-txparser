// Package parser contains the block poller and parsing logic.
package parser

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/danieloluwadare/tw-txparser/pkg/rpc"
	"github.com/danieloluwadare/tw-txparser/pkg/transaction"
)

// Start launches the polling goroutine if not already running.
func (p *parserImpl) Start(ctx context.Context) {
	p.pollingStartedMu.Lock()
	defer p.pollingStartedMu.Unlock()
	if p.pollingStarted {
		return
	}
	p.pollingStarted = true

	p.wg.Add(1)
	go p.pollLoop(ctx)
}

// Stop gracefully stops all goroutines and waits for them to complete.
func (p *parserImpl) Stop() {
	log.Println("[parser] stopping parser and waiting for goroutines to complete...")
	p.wg.Wait()
	log.Println("[parser] all goroutines stopped")
}

// pollLoop initializes the current block, kicks off scans, and runs forward scanning until cancelled.
func (p *parserImpl) pollLoop(ctx context.Context) {
	// Ensure pollingStarted flag is reset and WaitGroup is decremented when we exit
	defer func() {
		p.pollingStartedMu.Lock()
		p.pollingStarted = false
		p.pollingStartedMu.Unlock()
		p.wg.Done()
	}()
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	// --- Step 1: Initialize current block ---
	var blockHex string
	if err := p.client.Call(ctx, "eth_blockNumber", []interface{}{}, &blockHex); err != nil {
		log.Printf("[poll] failed to init current block: %v", err)
		return
	}
	latestBlock := hexToInt(blockHex)
	log.Printf("[poll] initialized at block %d", latestBlock)
	// --- Step 2: Process the latest block immediately ---
	if err := p.processBlock(ctx, latestBlock); err != nil {
		log.Printf("[poll] failed to process initial block %d: %v", latestBlock, err)
	}
	p.block = latestBlock

	// --- Step 3: Optionally start bounded backward scan in a goroutine ---
	if p.backwardScanEnabled {
		stopAt := latestBlock - p.backwardScanDepth
		if stopAt < 1 {
			stopAt = 1
		}
		p.wg.Add(1)
		go p.scanBackward(ctx, latestBlock-1, stopAt)
	}

	// --- Step 4: Forward scanning loop ---
	p.scanForward(ctx, ticker)
}

// scanBackward iterates from `from` down to `stopAt` (inclusive), processing each block.
func (p *parserImpl) scanBackward(ctx context.Context, from int, stopAt int) {
	defer p.wg.Done()
	log.Printf("[backward] starting scan from %d -> %d", from, stopAt)
	for i := from; i >= stopAt; i-- {
		select {
		case <-ctx.Done():
			log.Println("[backward] stopping backward scan")
			return
		default:
			if err := p.processBlock(ctx, i); err != nil {
				log.Printf("[backward] failed to process block %d: %v", i, err)
			}
			if i%1000 == 0 {
				log.Printf("[backward] scanned down to block %d", i)
			}
		}
	}
	log.Println("[backward] completed bounded historical scan")
}

// scanForward periodically checks for new blocks and processes them.
func (p *parserImpl) scanForward(ctx context.Context, ticker *time.Ticker) {
	log.Printf("[Forward] starting scan from %d ", p.block)
	for {
		select {
		case <-ctx.Done():
			log.Println("[forward] stopping forward scan")
			return
		case <-ticker.C:
			if err := p.checkForNewBlocks(ctx); err != nil {
				log.Printf("[forward] error checking new blocks: %v", err)
			}
		}
	}
}

// checkForNewBlocks queries the latest block number and processes newly discovered blocks.
func (p *parserImpl) checkForNewBlocks(ctx context.Context) error {
	var blockHex string
	if err := p.client.Call(ctx, "eth_blockNumber", []interface{}{}, &blockHex); err != nil {
		return fmt.Errorf("failed to get latest block number: %w", err)
	}
	latestBlock := hexToInt(blockHex)

	if latestBlock > p.block {
		for i := p.block + 1; i <= latestBlock; i++ {
			if err := p.processBlock(ctx, i); err != nil {
				log.Printf("[forward] failed to process block %d: %v", i, err)
			} else {
				log.Printf("[forward] processed block %d", i)
			}
		}
		p.block = latestBlock
	}
	return nil
}

// processBlock fetches a block by number and stores all transactions.
// Transactions are stored for both sender and receiver addresses, regardless of subscription status.
// This ensures no historical data is lost when addresses subscribe later.
func (p *parserImpl) processBlock(ctx context.Context, number int) error {
	var block rpc.Block
	if err := p.client.Call(ctx, "eth_getBlockByNumber", []interface{}{formatBlockNum(number), true}, &block); err != nil {
		return fmt.Errorf("failed to fetch block %d: %w", number, err)
	}

	for _, tx := range block.Transactions {
		log.Printf("to address: %s and from address: %s", tx.To, tx.From)

		// Store transaction for sender address (outbound from sender's perspective)
		p.store.AddTransaction(tx.From, transaction.Transaction{
			Hash:    tx.Hash,
			From:    tx.From,
			To:      tx.To,
			Value:   hexToBigIntString(tx.Value),
			Block:   number,
			Inbound: false, // Outbound transaction (from sender's perspective)
		})

		// Store transaction for receiver address (inbound from receiver's perspective)
		p.store.AddTransaction(tx.To, transaction.Transaction{
			Hash:    tx.Hash,
			From:    tx.From,
			To:      tx.To,
			Value:   hexToBigIntString(tx.Value),
			Block:   number,
			Inbound: true, // Inbound transaction (to receiver's perspective)
		})
	}
	return nil
}

// formatBlockNum converts a decimal block number into a 0x-prefixed hex string.
func formatBlockNum(num int) string {
	return "0x" + strconv.FormatInt(int64(num), 16)
}
