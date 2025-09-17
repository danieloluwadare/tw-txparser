// Package rpc provides a minimal JSON-RPC client and Ethereum types.
package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a simple JSON-RPC HTTP client.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a Client targeting the given RPC endpoint URL.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Call performs a JSON-RPC request and unmarshals the result into result.
func (c *Client) Call(ctx context.Context, method string, params []interface{}, result interface{}) error {
	req := JSONRPCRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("RPC call failed for method %s: %w", method, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RPC call failed with status %d for method %s", resp.StatusCode, method)
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to decode JSON-RPC response for method %s: %w", method, err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error for method %s (code %d): %s", method, rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if err := json.Unmarshal(rpcResp.Result, result); err != nil {
		return fmt.Errorf("failed to unmarshal result for method %s: %w", method, err)
	}
	return nil
}

// GetBlockNumber returns the latest block number as a hex string.
func (c *Client) GetBlockNumber(ctx context.Context) (string, error) {
	var blockHex string
	err := c.Call(ctx, "eth_blockNumber", []interface{}{}, &blockHex)
	if err != nil {
		return "", fmt.Errorf("failed to get block number: %w", err)
	}
	return blockHex, nil
}

// GetBlockByNumber returns block details for the given block number.
// blockNumber should be a hex string (e.g., "0x1234" or "latest").
// includeTransactions determines whether to include full transaction objects.
func (c *Client) GetBlockByNumber(ctx context.Context, blockNumber string, includeTransactions bool) (*Block, error) {
	var block Block
	err := c.Call(ctx, "eth_getBlockByNumber", []interface{}{blockNumber, includeTransactions}, &block)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", blockNumber, err)
	}
	return &block, nil
}

// GetBlockByNumberInt returns block details for the given block number as an integer.
// This is a convenience method that converts the integer to hex format.
func (c *Client) GetBlockByNumberInt(ctx context.Context, blockNumber int, includeTransactions bool) (*Block, error) {
	hexBlockNumber := fmt.Sprintf("0x%x", blockNumber)
	return c.GetBlockByNumber(ctx, hexBlockNumber, includeTransactions)
}
