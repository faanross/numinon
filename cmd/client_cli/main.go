package main

import (
	"encoding/json"
	"flag"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"numinon_shadow/internal/clientapi"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	defaultServerWsURL = "ws://localhost:8080/client"
)

func main() {
	// Define flags for global settings
	serverURL := flag.String("server", defaultServerWsURL, "C2 server WebSocket API URL")

	// Define flags for actions and their parameters
	action := flag.String("action", "", "Action to perform: create, list, stop")

	// Flags for 'create' action
	createProto := flag.String("proto", "H1C", "Listener protocol (H1C, H1TLS, WSS, etc.) for create action")
	createAddr := flag.String("addr", ":7777", "Listener address (e.g., 0.0.0.0:7777 or :7777) for create action")
	createCertPath := flag.String("cert", "", "Path to TLS certificate (for H1TLS, WSS, etc.) for create action")
	createKeyPath := flag.String("key", "", "Path to TLS private key (for H1TLS, WSS, etc.) for create action")

	// Flags for 'stop' action
	stopListenerID := flag.String("id", "", "Listener ID to stop for stop action")

	flag.Parse()

	if *action == "" {
		log.Println("Error: -action flag is required (create, list, or stop).")
		flag.Usage()
		os.Exit(1)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u, err := url.Parse(*serverURL)
	if err != nil {
		log.Fatalf("Error parsing server URL: %v", err)
	}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Dial error to %s: %v", u.String(), err)
	}
	defer conn.Close()

	log.Println("Connected to server.")

	done := make(chan struct{})
	go func() {
		defer close(done) // This will signal the main function to proceed after this goroutine exits
		// Read just ONE message
		_, message, errRead := conn.ReadMessage()
		if errRead != nil {
			log.Printf("Read error: %v (Connection might be closed)", errRead)
			return // Exit goroutine, close(done) will be called
		}
		log.Printf("Received from server: %s\n\n", string(message))
		// After processing one message, this goroutine will exit.
		// The 'defer close(done)' will then be executed.
	}()

	// Construct and send the request based on action
	var opRequest clientapi.ClientRequest
	opRequest.RequestID = "cli_req_" + uuid.NewString()[:8] // Generate a unique request ID

	switch strings.ToLower(*action) {
	case "create":
		opRequest.Action = clientapi.ActionCreateListener
		payload := clientapi.CreateListenerPayload{
			Protocol: *createProto,
			Address:  *createAddr,
			CertPath: *createCertPath,
			KeyPath:  *createKeyPath,
		}
		payloadBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			log.Fatalf("Failed to marshal create listener payload: %v", marshalErr)
		}
		opRequest.Payload = payloadBytes
		log.Printf("Sending CreateListener request: %+v", payload)

	case "list":
		opRequest.Action = clientapi.ActionListListeners
		// Payload for list is typically empty or can be an empty JSON object
		opRequest.Payload = json.RawMessage("{}")
		log.Println("Sending ListListeners request...")

	case "stop":
		opRequest.Action = clientapi.ActionStopListener
		if *stopListenerID == "" {
			log.Fatal("Error: -id flag is required for 'stop' action.")
		}
		payload := clientapi.StopListenerPayload{
			ListenerID: *stopListenerID,
		}
		payloadBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			log.Fatalf("Failed to marshal stop listener payload: %v", marshalErr)
		}
		opRequest.Payload = payloadBytes
		log.Printf("Sending StopListener request for ID: %s", *stopListenerID)

	default:
		log.Fatalf("Unknown action: %s. Valid actions are create, list, stop.", *action)
	}

	reqBytes, err := json.Marshal(opRequest)
	if err != nil {
		log.Fatalf("Failed to marshal OperatorRequest: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, reqBytes)
	if err != nil {
		log.Fatalf("Write error: %v", err)
	}
	log.Printf("Sent request to server: %s", string(reqBytes))

	// Wait for a response or interrupt
	select {
	case <-done:
		log.Println("Read loop finished (connection closed by server or error).")
	case <-time.After(10 * time.Second): // Timeout for waiting for a response
		log.Println("Timeout waiting for server response.")
	case <-interrupt:
		log.Println("Interrupt received, closing connection.")
		// Cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Write close error:", err)
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
	log.Println("Operator CLI finished.")
}
