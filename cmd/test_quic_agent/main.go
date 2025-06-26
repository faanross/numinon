package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

func main() {
	// --- Configuration ---

	const targetURL = "https://127.0.0.1:8888/"
	log.Printf("Targeting HTTP/3 server at: %s", targetURL)

	log.Println("Setting up HTTP/3 client...")

	transport := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Allow self-signed certificates for local dev
		},
		// QUICConfig can be used for more fine-grained QUIC protocol settings.
		// For this simple client, the defaults are sufficient.
	}

	// Create the main HTTP client, passing in our custom HTTP/3 transport.
	client := &http.Client{
		Transport: transport,
	}

	// Ensure the Transport is closed gracefully to release underlying connections.
	defer transport.Close()

	// --- Make Request ---
	log.Println("Sending GET request...")
	resp, err := client.Get(targetURL)
	if err != nil {
		log.Fatalf("|❗ERR| Failed to perform GET request: %v", err)
	}
	// Ensure the response body is closed even if errors occur later
	defer resp.Body.Close()

	// --- Process Response ---
	log.Printf("Received response:")
	log.Printf("  Status Code: %d (%s)", resp.StatusCode, resp.Status)
	log.Printf("  Protocol: %s", resp.Proto) // Should indicate HTTP/3.0

	// Check if the protocol is indeed HTTP/3
	if resp.ProtoMajor != 3 {
		log.Printf("|⚠️ WARN| Expected HTTP/3, but got %s", resp.Proto)
	} else {
		log.Println("  Successfully connected using HTTP/3!")
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Still log status code etc. even if body reading fails
		log.Printf("|❗ERR| Failed to read response body: %v", err)
	}

	log.Printf("  Response Body:\n---\n%s\n---", string(bodyBytes))

	if resp.StatusCode == http.StatusOK {
		log.Println("|✅ OK | Request successful.")
	} else {
		log.Printf("|⚠️ WARN| Request completed but received non-200 status: %d", resp.StatusCode)
	}
}
