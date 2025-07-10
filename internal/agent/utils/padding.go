package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

// GenerateRandomPadding creates a random base64 encoded string for padding.
// The length of the raw random bytes will be between minRawBytes and maxRawBytes (inclusive).
// If maxRawBytes is 0 or less, no padding is generated, and an empty string is returned.
// If minRawBytes < 0, it's treated as 0.
// If minRawBytes > maxRawBytes, minRawBytes is used as maxRawBytes (effectively generating padding of size minRawBytes).
func GenerateRandomPadding(minRawBytes, maxRawBytes int) (string, error) {
	if maxRawBytes <= 0 {
		return "", nil // No padding requested or possible
	}
	if minRawBytes < 0 {
		minRawBytes = 0
	}
	if minRawBytes > maxRawBytes {
		// If min is greater, effectively just generate padding of size minRawBytes + 1000
		// (or you could error, but this is more forgiving)
		maxRawBytes = minRawBytes + 1000
	}

	numBytesToGenerate := minRawBytes

	if maxRawBytes > minRawBytes {
		rangeBytes := maxRawBytes - minRawBytes + 1 // +1 to make maxRawBytes inclusive
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(rangeBytes)))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number for padding size: %w", err)
		}
		numBytesToGenerate += int(nBig.Int64())
	}

	if numBytesToGenerate <= 0 { // Double check if calculations resulted in zero or less
		return "", nil
	}

	randomBytes := make([]byte, numBytesToGenerate)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes for padding: %w", err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}
