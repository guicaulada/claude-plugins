package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

// TraceID generates a 32-character hex string for use as an OTel trace ID.
func TraceID() string {
	return randomHex(16)
}

// SpanID generates a 16-character hex string for use as an OTel span ID.
func SpanID() string {
	return randomHex(8)
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		// Fallback should never happen with crypto/rand
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
