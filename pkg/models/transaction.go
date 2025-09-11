// Package models defines shared domain models.
package models

// Transaction is a normalized transaction persisted per address.
type Transaction struct {
	Hash    string `json:"hash"`
	From    string `json:"from"`
	To      string `json:"to"`
	Value   string `json:"value"`
	Block   int    `json:"block"`
	Inbound bool   `json:"inbound"` // true if transaction is TO the subscribed address
}
