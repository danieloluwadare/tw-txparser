// Package parser contains the block poller and parsing logic.
package parser

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
)

// hexToInt parses a hex string (with or without 0x prefix) into int.
func hexToInt(hexStr string) int {
	val, _ := strconv.ParseInt(strings.TrimPrefix(hexStr, "0x"), 16, 64)
	return int(val)
}

// decodeHex decodes a hex string into its first byte value.
// Returns an error for empty or invalid input.
func decodeHex(hexStr string) (int, error) {
	trimmed := strings.TrimPrefix(hexStr, "0x")
	if trimmed == "" {
		return 0, hex.InvalidByteError(0)
	}
	// Pad odd-length hex strings with leading zero
	if len(trimmed)%2 == 1 {
		trimmed = "0" + trimmed
	}
	b, err := hex.DecodeString(trimmed)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	return int(b[0]), nil
}

// hexToBigIntString converts hex string "0x..." to decimal string.
// Values larger than 256 bits are truncated to the least-significant 256 bits.
func hexToBigIntString(h string) string {
	hi := strings.TrimPrefix(h, "0x")
	if hi == "" {
		return "0"
	}
	// Cap to 256 bits (64 hex chars) by taking the least-significant digits
	if len(hi) > 64 {
		hi = hi[len(hi)-64:]
	}
	b := new(big.Int)
	_, ok := b.SetString(hi, 16)
	if !ok {
		return "0"
	}
	return b.String()
}
