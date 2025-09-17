// Package rpc provides a minimal JSON-RPC client and Ethereum types.
package rpc

import (
	"context"
	"encoding/json"
)

// RPCClient abstracts a JSON-RPC caller.
type RPCClient interface {
	Call(ctx context.Context, method string, params []interface{}, result interface{}) error
}

// JSONRPCRequest is the wire format for requests.
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSONRPCResponse is the wire format for responses.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError models an error object in JSON-RPC responses.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error satisfies the error interface.
func (e *RPCError) Error() string {
	return e.Message
}

// Block describes an Ethereum block with basic fields used by this app.
type Block struct {
	Number       string        `json:"number"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction describes an Ethereum transaction in RPC responses.
type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}
