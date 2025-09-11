// Package rpc provides a minimal JSON-RPC client and Ethereum types.
package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client is a simple JSON-RPC HTTP client.
type Client struct {
	endpoint string
}

// NewClient creates a Client targeting the given RPC endpoint URL.
func NewClient(endpoint string) *Client {
	return &Client{endpoint: endpoint}
}

// Call performs a JSON-RPC request and unmarshals the result into result.
func (c *Client) Call(method string, params []interface{}, result interface{}) error {
	req := JSONRPCRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1}
	body, _ := json.Marshal(req)

	resp, err := http.Post(c.endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error: %s", rpcResp.Error.Message)
	}
	return json.Unmarshal(rpcResp.Result, result)
}
