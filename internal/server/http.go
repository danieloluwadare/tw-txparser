// Package server exposes HTTP endpoints for subscription and transaction queries.
package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/danieloluwadare/tw-txparser/pkg/parser"
)

// Server hosts HTTP handlers that proxy to a parser.Parser.
type Server struct {
	parser parser.Parser
}

// New constructs a Server with the provided parser.
func New(p parser.Parser) *Server {
	return &Server{parser: p}
}

// Start binds handlers and starts listening on addr.
func (s *Server) Start(addr string) error {
	http.HandleFunc("/subscribe", s.HandleSubscribe)
	http.HandleFunc("/current", s.HandleCurrentBlock)
	http.HandleFunc("/transactions", s.HandleTransactions)
	return http.ListenAndServe(addr, nil)
}

// HandleSubscribe subscribes an address via POST {"address":"..."}.
func (s *Server) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Address == "" {
		http.Error(w, "missing address", http.StatusBadRequest)
		return
	}

	ok := s.parser.Subscribe(body.Address)
	if err := json.NewEncoder(w).Encode(map[string]bool{"subscribed": ok}); err != nil {
		log.Println("failed to encode response:", err)
	}
}

// HandleCurrentBlock returns the latest known block as {"block":N}.
func (s *Server) HandleCurrentBlock(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]int{"block": s.parser.GetCurrentBlock()})
}

// HandleTransactions returns transactions associated with a given address query param.
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("address")
	if addr == "" {
		http.Error(w, "missing address", http.StatusBadRequest)
		return
	}
	txs := s.parser.GetTransactions(addr)
	if err := json.NewEncoder(w).Encode(txs); err != nil {
		log.Println("failed to encode response:", err)
	}
}
