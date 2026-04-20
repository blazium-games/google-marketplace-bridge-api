package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateConsumerAdminURLs returns two distinct random HTTPS URLs per call.
// consumer_url is associated with the /dummy flow; admin_url with /instantiate (see doc).
func GenerateConsumerAdminURLs() (consumerURL, adminURL string) {
	var a, b [12]byte
	_, _ = rand.Read(a[:])
	_, _ = rand.Read(b[:])
	consumerURL = fmt.Sprintf("https://consumer.pogr-bridge.invalid/%s", hex.EncodeToString(a[:]))
	adminURL = fmt.Sprintf("https://admin.pogr-bridge.invalid/%s", hex.EncodeToString(b[:]))
	return consumerURL, adminURL
}
