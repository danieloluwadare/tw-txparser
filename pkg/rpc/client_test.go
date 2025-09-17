package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Call(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Verify request structure
		if req.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC 2.0, got %s", req.JSONRPC)
		}
		if req.ID != 1 {
			t.Errorf("Expected ID 1, got %d", req.ID)
		}

		// Send response based on method
		var response JSONRPCResponse
		switch req.Method {
		case "eth_blockNumber":
			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`"0x1234"`),
			}
		case "eth_getBlockByNumber":
			block := Block{
				Number: "0x1234",
				Transactions: []Transaction{
					{
						Hash:  "0xhash1",
						From:  "0xfrom1",
						To:    "0xto1",
						Value: "0x1000",
					},
				},
			}
			blockJSON, _ := json.Marshal(block)
			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  blockJSON,
			}
		default:
			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`"test"`),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL)

	// Test eth_blockNumber call
	var blockNumber string
	err := client.Call(context.Background(), "eth_blockNumber", []interface{}{}, &blockNumber)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if blockNumber != "0x1234" {
		t.Errorf("Expected block number 0x1234, got %s", blockNumber)
	}

	// Test eth_getBlockByNumber call
	var block Block
	err = client.Call(context.Background(), "eth_getBlockByNumber", []interface{}{"0x1234", true}, &block)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if block.Number != "0x1234" {
		t.Errorf("Expected block number 0x1234, got %s", block.Number)
	}
	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
	}
	if block.Transactions[0].Hash != "0xhash1" {
		t.Errorf("Expected transaction hash 0xhash1, got %s", block.Transactions[0].Hash)
	}
}

func TestClient_Call_Error(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result string
	err := client.Call(context.Background(), "invalid_method", []interface{}{}, &result)
	if err == nil {
		t.Error("Expected error but got none")
	}
	expectedError := "RPC error for method invalid_method (code -32601): Method not found"
	if err.Error() != expectedError {
		t.Errorf("Expected '%s', got %s", expectedError, err.Error())
	}
}

func TestClient_Call_NetworkError(t *testing.T) {
	// Create client with invalid URL
	client := NewClient("http://invalid-url-that-does-not-exist")

	var result string
	err := client.Call(context.Background(), "eth_blockNumber", []interface{}{}, &result)
	if err == nil {
		t.Error("Expected network error but got none")
	}
}

func TestClient_Call_InvalidJSON(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result string
	err := client.Call(context.Background(), "eth_blockNumber", []interface{}{}, &result)
	if err == nil {
		t.Error("Expected JSON decode error but got none")
	}
}

func TestJSONRPCRequest(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "test_method",
		Params:  []interface{}{"param1", 123},
		ID:      1,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledReq JSONRPCRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// Verify all fields
	if unmarshaledReq.JSONRPC != req.JSONRPC {
		t.Errorf("JSONRPC mismatch: got %s, expected %s", unmarshaledReq.JSONRPC, req.JSONRPC)
	}
	if unmarshaledReq.Method != req.Method {
		t.Errorf("Method mismatch: got %s, expected %s", unmarshaledReq.Method, req.Method)
	}
	if unmarshaledReq.ID != req.ID {
		t.Errorf("ID mismatch: got %d, expected %d", unmarshaledReq.ID, req.ID)
	}
}

func TestJSONRPCResponse(t *testing.T) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  json.RawMessage(`"test_result"`),
		Error:   nil,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledResp JSONRPCResponse
	err = json.Unmarshal(jsonData, &unmarshaledResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify all fields
	if unmarshaledResp.JSONRPC != response.JSONRPC {
		t.Errorf("JSONRPC mismatch: got %s, expected %s", unmarshaledResp.JSONRPC, response.JSONRPC)
	}
	if unmarshaledResp.ID != response.ID {
		t.Errorf("ID mismatch: got %d, expected %d", unmarshaledResp.ID, response.ID)
	}
	if string(unmarshaledResp.Result) != string(response.Result) {
		t.Errorf("Result mismatch: got %s, expected %s", string(unmarshaledResp.Result), string(response.Result))
	}
}

func TestRPCError(t *testing.T) {
	rpcError := RPCError{
		Code:    -32601,
		Message: "Method not found",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(rpcError)
	if err != nil {
		t.Fatalf("Failed to marshal RPC error: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledError RPCError
	err = json.Unmarshal(jsonData, &unmarshaledError)
	if err != nil {
		t.Fatalf("Failed to unmarshal RPC error: %v", err)
	}

	// Verify all fields
	if unmarshaledError.Code != rpcError.Code {
		t.Errorf("Code mismatch: got %d, expected %d", unmarshaledError.Code, rpcError.Code)
	}
	if unmarshaledError.Message != rpcError.Message {
		t.Errorf("Message mismatch: got %s, expected %s", unmarshaledError.Message, rpcError.Message)
	}
}

func TestClient_GetBlockNumber(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"jsonrpc":"2.0","id":1,"result":"0x1234"}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Test GetBlockNumber
	blockNumber, err := client.GetBlockNumber(context.Background())
	if err != nil {
		t.Fatalf("GetBlockNumber failed: %v", err)
	}
	if blockNumber != "0x1234" {
		t.Errorf("Expected block number 0x1234, got %s", blockNumber)
	}
}

func TestClient_GetBlockByNumber(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"jsonrpc":"2.0","id":1,"result":{"number":"0x1234","hash":"0xabcd","transactions":[]}}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Test GetBlockByNumber with hex string
	block, err := client.GetBlockByNumber(context.Background(), "0x1234", true)
	if err != nil {
		t.Fatalf("GetBlockByNumber failed: %v", err)
	}
	if block.Number != "0x1234" {
		t.Errorf("Expected block number 0x1234, got %s", block.Number)
	}
}

func TestClient_GetBlockByNumberInt(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"jsonrpc":"2.0","id":1,"result":{"number":"0x1234","hash":"0xabcd","transactions":[]}}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Test GetBlockByNumberInt with integer
	block, err := client.GetBlockByNumberInt(context.Background(), 4660, true)
	if err != nil {
		t.Fatalf("GetBlockByNumberInt failed: %v", err)
	}
	if block.Number != "0x1234" {
		t.Errorf("Expected block number 0x1234, got %s", block.Number)
	}
}
