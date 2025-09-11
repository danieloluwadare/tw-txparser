package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danieloluwadare/tw-txparser/pkg/models"
)

// MockParser implements the parser.Parser interface for testing
type MockParser struct {
	currentBlock  int
	transactions  map[string][]models.Transaction
	subscriptions map[string]bool
}

func NewMockParser() *MockParser {
	return &MockParser{
		transactions:  make(map[string][]models.Transaction),
		subscriptions: make(map[string]bool),
	}
}

func (m *MockParser) GetCurrentBlock() int {
	return m.currentBlock
}

func (m *MockParser) Subscribe(address string) bool {
	if m.subscriptions[address] {
		return false
	}
	m.subscriptions[address] = true
	return true
}

func (m *MockParser) GetTransactions(address string) []models.Transaction {
	return m.transactions[address]
}

func TestServer_New(t *testing.T) {
	parser := NewMockParser()
	server := New(parser)

	if server == nil {
		t.Fatal("Expected server to be created")
	}
	if server.parser != parser {
		t.Error("Expected server to use the provided parser")
	}
}

func TestServer_HandleSubscribe(t *testing.T) {
	parser := NewMockParser()
	server := New(parser)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectedBody   map[string]bool
	}{
		{
			name:   "successful subscription",
			method: http.MethodPost,
			body: map[string]string{
				"address": "0x1234567890abcdef",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]bool{"subscribed": true},
		},
		{
			name:   "duplicate subscription",
			method: http.MethodPost,
			body: map[string]string{
				"address": "0x1234567890abcdef",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]bool{"subscribed": false},
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name:   "missing address",
			method: http.MethodPost,
			body: map[string]string{
				"address": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/subscribe", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.HandleSubscribe(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != nil {
				var response map[string]bool
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if response["subscribed"] != tt.expectedBody["subscribed"] {
					t.Errorf("Expected subscribed %t, got %t", tt.expectedBody["subscribed"], response["subscribed"])
				}
			}
		})
	}
}

func TestServer_HandleCurrentBlock(t *testing.T) {
	parser := NewMockParser()
	parser.currentBlock = 12345
	server := New(parser)

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	w := httptest.NewRecorder()

	server.HandleCurrentBlock(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]int
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["block"] != 12345 {
		t.Errorf("Expected block 12345, got %d", response["block"])
	}
}

func TestServer_HandleTransactions(t *testing.T) {
	parser := NewMockParser()
	server := New(parser)

	// Add some test transactions
	address := "0x1234567890abcdef"
	transactions := []models.Transaction{
		{Hash: "0xhash1", From: "0xfrom1", To: address, Value: "1000", Block: 1, Inbound: true},
		{Hash: "0xhash2", From: "0xfrom2", To: address, Value: "2000", Block: 2, Inbound: true},
	}
	parser.transactions[address] = transactions

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "valid address",
			queryParams:    "?address=0x1234567890abcdef",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "missing address",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "empty address",
			queryParams:    "?address=",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "non-existent address",
			queryParams:    "?address=0xnonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/transactions"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			server.HandleTransactions(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response []models.Transaction
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(response) != tt.expectedCount {
					t.Errorf("Expected %d transactions, got %d", tt.expectedCount, len(response))
				}
			}
		})
	}
}

func TestServer_Start(t *testing.T) {
	parser := NewMockParser()
	server := New(parser)

	// Test that Start method exists and can be called
	// Note: We can't easily test the actual HTTP server without mocking net.Listen
	// This test just ensures the method exists and doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Start method panicked: %v", r)
		}
	}()

	// We expect this to fail because we're not providing a valid address
	// but it should not panic
	err := server.Start("invalid-address")
	if err == nil {
		t.Error("Expected Start to return an error for invalid address")
	}
}

func TestServer_Integration(t *testing.T) {
	parser := NewMockParser()
	server := New(parser)

	// Test full flow: subscribe -> get current block -> get transactions
	address := "0x1234567890abcdef"

	// 1. Subscribe to address
	subscribeBody := map[string]string{"address": address}
	body, _ := json.Marshal(subscribeBody)
	req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleSubscribe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Subscribe failed with status %d", w.Code)
	}

	// 2. Get current block
	req = httptest.NewRequest(http.MethodGet, "/current", nil)
	w = httptest.NewRecorder()
	server.HandleCurrentBlock(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get current block failed with status %d", w.Code)
	}

	// 3. Get transactions (should be empty initially)
	req = httptest.NewRequest(http.MethodGet, "/transactions?address="+address, nil)
	w = httptest.NewRecorder()
	server.HandleTransactions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get transactions failed with status %d", w.Code)
	}

	var transactions []models.Transaction
	if err := json.NewDecoder(w.Body).Decode(&transactions); err != nil {
		t.Fatalf("Failed to decode transactions: %v", err)
	}

	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions initially, got %d", len(transactions))
	}
}

func TestServer_ErrorHandling(t *testing.T) {
	parser := NewMockParser()
	server := New(parser)

	// Test handling of malformed JSON in subscribe request
	req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleSubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}
