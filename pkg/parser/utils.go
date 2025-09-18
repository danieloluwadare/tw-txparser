// Package parser contains the block poller and parsing logic.
package parser

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
)

// hexToInt parses a hex string (with or without 0x prefix) into int.
// Returns 0 if parsing fails.
func hexToInt(hexStr string) int {
	val, err := strconv.ParseInt(strings.TrimPrefix(hexStr, "0x"), 16, 64)
	if err != nil {
		// Log the error but don't fail the entire operation
		// This is used in polling where we want to continue even if one block fails
		return 0
	}
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
// Returns "0" if parsing fails.
func hexToBigIntString(h string) string {
	hi := strings.TrimPrefix(h, "0x")
	if hi == "" {
		return "0"
	}
	b := new(big.Int)
	_, ok := b.SetString(hi, 16)
	if !ok {
		// Return "0" for invalid hex strings rather than failing
		// This ensures the parser continues even with malformed data
		return "0"
	}
	return b.String()
}
