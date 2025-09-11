package models

import (
	"encoding/json"
	"testing"
)

func TestTransaction(t *testing.T) {
	tx := Transaction{
		Hash:  "0x1234567890abcdef",
		From:  "0xfromaddress",
		To:    "0xtoaddress",
		Value: "1000000000000000000",
		Block: 12345,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledTx Transaction
	err = json.Unmarshal(jsonData, &unmarshaledTx)
	if err != nil {
		t.Fatalf("Failed to unmarshal transaction: %v", err)
	}

	// Verify all fields are preserved
	if unmarshaledTx.Hash != tx.Hash {
		t.Errorf("Hash mismatch: got %s, expected %s", unmarshaledTx.Hash, tx.Hash)
	}
	if unmarshaledTx.From != tx.From {
		t.Errorf("From mismatch: got %s, expected %s", unmarshaledTx.From, tx.From)
	}
	if unmarshaledTx.To != tx.To {
		t.Errorf("To mismatch: got %s, expected %s", unmarshaledTx.To, tx.To)
	}
	if unmarshaledTx.Value != tx.Value {
		t.Errorf("Value mismatch: got %s, expected %s", unmarshaledTx.Value, tx.Value)
	}
	if unmarshaledTx.Block != tx.Block {
		t.Errorf("Block mismatch: got %d, expected %d", unmarshaledTx.Block, tx.Block)
	}
}
