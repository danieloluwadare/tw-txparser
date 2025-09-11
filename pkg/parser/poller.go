// Package parser contains the block poller and parsing logic.
package parser

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/danieloluwadare/tw-txparser/pkg/models"
	"github.com/danieloluwadare/tw-txparser/pkg/rpc"
)

// Start launches the polling goroutine if not already running.
func (p *parserImpl) Start(ctx context.Context) {
	p.pollingStartedMu.Lock()
	defer p.pollingStartedMu.Unlock()
	if p.pollingStarted {
		return
	}
	p.pollingStarted = true

	go p.pollLoop(ctx)
}

// pollLoop initializes the current block, kicks off scans, and runs forward scanning until cancelled.
func (p *parserImpl) pollLoop(ctx context.Context) {
	// Ensure pollingStarted flag is reset when we exit
	defer func() {
		p.pollingStartedMu.Lock()
		p.pollingStarted = false
		p.pollingStartedMu.Unlock()
	}()
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	// --- Step 1: Initialize current block ---
	var blockHex string
	if err := p.client.Call("eth_blockNumber", []interface{}{}, &blockHex); err != nil {
		log.Printf("[poll] failed to init current block: %v", err)
		return
	}
	latestBlock := hexToInt(blockHex)
	log.Printf("[poll] initialized at block %d", latestBlock)
	// --- Step 2: Process the latest block immediately ---
	p.processBlock(latestBlock)
	p.block = latestBlock

	// --- Step 3: Optionally start bounded backward scan in a goroutine ---
	if p.backwardScanEnabled {
		stopAt := latestBlock - p.backwardScanDepth
		if stopAt < 1 {
			stopAt = 1
		}
		go p.scanBackward(ctx, latestBlock-1, stopAt)
	}

	// --- Step 4: Forward scanning loop ---
	p.scanForward(ctx, ticker)
}

// scanBackward iterates from `from` down to `stopAt` (inclusive), processing each block.
func (p *parserImpl) scanBackward(ctx context.Context, from int, stopAt int) {
	log.Printf("[backward] starting scan from %d -> %d", from, stopAt)
	for i := from; i >= stopAt; i-- {
		select {
		case <-ctx.Done():
			log.Println("[backward] stopping backward scan")
			return
		default:
			p.processBlock(i)
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
			if err := p.checkForNewBlocks(); err != nil {
				log.Printf("[forward] error checking new blocks: %v", err)
			}
		}
	}
}

// checkForNewBlocks queries the latest block number and processes newly discovered blocks.
func (p *parserImpl) checkForNewBlocks() error {
	var blockHex string
	if err := p.client.Call("eth_blockNumber", []interface{}{}, &blockHex); err != nil {
		return err
	}
	latestBlock := hexToInt(blockHex)

	if latestBlock > p.block {
		for i := p.block + 1; i <= latestBlock; i++ {
			p.processBlock(i)
			log.Printf("[forward] processed block %d", i)
		}
		p.block = latestBlock
	}
	return nil
}

// processBlock fetches a block by number and stores transactions for sender and receiver addresses.
func (p *parserImpl) processBlock(number int) {
	var block rpc.Block
	if err := p.client.Call("eth_getBlockByNumber", []interface{}{formatBlockNum(number), true}, &block); err != nil {
		log.Println("error fetching block:", err)
		return
	}

	for _, tx := range block.Transactions {
		log.Printf("from address %v to address %v", tx.From, tx.To)
		p.store.AddTransaction(tx.From, models.Transaction{
			Hash:  tx.Hash,
			From:  tx.From,
			To:    tx.To,
			Value: hexToBigIntString(tx.Value),
			Block: number,
		})
		p.store.AddTransaction(tx.To, models.Transaction{
			Hash:  tx.Hash,
			From:  tx.From,
			To:    tx.To,
			Value: hexToBigIntString(tx.Value),
			Block: number,
		})
	}
}

// formatBlockNum converts a decimal block number into a 0x-prefixed hex string.
func formatBlockNum(num int) string {
	return "0x" + strconv.FormatInt(int64(num), 16)
}
