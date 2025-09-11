package parser

import (
	"testing"
)

func TestHexToInt(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected int
	}{
		{
			name:     "valid hex with 0x prefix",
			hexStr:   "0x1a",
			expected: 26,
		},
		{
			name:     "valid hex without 0x prefix",
			hexStr:   "1a",
			expected: 26,
		},
		{
			name:     "zero value",
			hexStr:   "0x0",
			expected: 0,
		},
		{
			name:     "large hex value",
			hexStr:   "0xffff",
			expected: 65535,
		},
		{
			name:     "empty string",
			hexStr:   "",
			expected: 0,
		},
		{
			name:     "invalid hex",
			hexStr:   "0xgg",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexToInt(tt.hexStr)
			if result != tt.expected {
				t.Errorf("hexToInt(%s) = %d, expected %d", tt.hexStr, result, tt.expected)
			}
		})
	}
}

func TestDecodeHex(t *testing.T) {
	tests := []struct {
		name        string
		hexStr      string
		expected    int
		expectError bool
	}{
		{
			name:        "valid hex with 0x prefix",
			hexStr:      "0x1a",
			expected:    26,
			expectError: false,
		},
		{
			name:        "valid hex without 0x prefix",
			hexStr:      "1a",
			expected:    26,
			expectError: false,
		},
		{
			name:        "zero value with padding",
			hexStr:      "0x00",
			expected:    0,
			expectError: false,
		},
		{
			name:        "zero value single digit",
			hexStr:      "0x0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "empty string",
			hexStr:      "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid hex",
			hexStr:      "0xgg",
			expected:    0,
			expectError: true,
		},
		{
			name:        "odd length hex",
			hexStr:      "0x1",
			expected:    1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeHex(tt.hexStr)
			if tt.expectError {
				if err == nil {
					t.Errorf("decodeHex(%s) expected error but got none", tt.hexStr)
				}
			} else {
				if err != nil {
					t.Errorf("decodeHex(%s) unexpected error: %v", tt.hexStr, err)
				}
				if result != tt.expected {
					t.Errorf("decodeHex(%s) = %d, expected %d", tt.hexStr, result, tt.expected)
				}
			}
		})
	}
}

func TestHexToBigIntString(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected string
	}{
		{
			name:     "valid hex with 0x prefix",
			hexStr:   "0x1a",
			expected: "26",
		},
		{
			name:     "valid hex without 0x prefix",
			hexStr:   "1a",
			expected: "26",
		},
		{
			name:     "zero value",
			hexStr:   "0x0",
			expected: "0",
		},
		{
			name:     "empty string",
			hexStr:   "",
			expected: "0",
		},
		{
			name:     "large hex value",
			hexStr:   "0xffffffffffffffff",
			expected: "18446744073709551615",
		},
		{
			name:     "very large hex value",
			hexStr:   "0x1ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			expected: "26815615859885194199148049996411692254958731641184786755447122887443528060147093953603748596333806855380063716372972101707507765623893139892867298012168191",
		},
		{
			name:     "invalid hex",
			hexStr:   "0xgg",
			expected: "0",
		},
		{
			name:     "hex with leading zeros",
			hexStr:   "0x0001a",
			expected: "26",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexToBigIntString(tt.hexStr)
			if result != tt.expected {
				t.Errorf("hexToBigIntString(%s) = %s, expected %s", tt.hexStr, result, tt.expected)
			}
		})
	}
}
