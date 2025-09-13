package listenermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/clientapi"
	"numinon_shadow/internal/listener"
	"strings"
)

// OperatorListenerManager implements clientapi.ListenerManager.
// It wraps the existing listener.Manager to provide operator-friendly functionality.
//
// Think of this as a "customer service representative" for the listener system:
// - It translates operator requests into system actions
// - It formats system responses for operator consumption
// - It adds operator-specific features (like session tracking)
type OperatorListenerManager struct {
	core *listener.Manager // The actual listener manager
}

// NewOperatorListenerManager creates a new operator-friendly wrapper.
func NewOperatorListenerManager(core *listener.Manager) *OperatorListenerManager {
	return &OperatorListenerManager{
		core: core,
	}
}

// CreateListener handles operator requests to create new listeners.
func (m *OperatorListenerManager) CreateListener(ctx context.Context, req clientapi.CreateListenerPayload, operatorSessionID string) (clientapi.ServerResponse, error) {
	log.Printf("[LISTENER API] Operator %s requesting new listener: %s on %s",
		operatorSessionID, req.Protocol, req.Address)

	// Step 1: Parse the address (operators provide "0.0.0.0:8080", we need IP and port separate)
	parts := strings.Split(req.Address, ":")
	if len(parts) != 2 {
		return clientapi.ServerResponse{
			Status: clientapi.StatusError,
			Error:  "Invalid address format. Expected format: IP:PORT (e.g., 0.0.0.0:8080)",
		}, fmt.Errorf("invalid address format: %s", req.Address)
	}

	ip := parts[0]
	port := parts[1]

	// Step 2: Map operator protocol names to internal types
	listenerType, err := m.mapProtocolToListenerType(req.Protocol)
	if err != nil {
		return clientapi.ServerResponse{
			Status: clientapi.StatusError,
			Error:  err.Error(),
		}, err
	}

	// Step 3: Create the listener using the core manager
	listenerID, err := m.core.CreateListener(listenerType, ip, port)
	if err != nil {
		return clientapi.ServerResponse{
			Status: clientapi.StatusError,
			Error:  fmt.Sprintf("Failed to create listener: %v", err),
		}, err
	}

	// Step 4: Format success response for operator
	statusPayload := clientapi.ListenerStatusPayload{
		ListenerID: listenerID,
		Protocol:   req.Protocol,
		Address:    req.Address,
		Status:     "RUNNING",
		Message:    fmt.Sprintf("Listener created successfully by operator %s", operatorSessionID),
	}

	payloadBytes, _ := json.Marshal(statusPayload)

	log.Printf("[LISTENER API] Successfully created listener %s for operator %s",
		listenerID, operatorSessionID)

	return clientapi.ServerResponse{
		Status:   clientapi.StatusSuccess,
		DataType: clientapi.DataTypeListenerStatus,
		Payload:  payloadBytes,
	}, nil
}

// ListListeners returns all active listeners in operator-friendly format.
func (m *OperatorListenerManager) ListListeners(ctx context.Context, operatorSessionID string) (clientapi.ServerResponse, error) {
	log.Printf("[LISTENER API] Operator %s requesting listener list", operatorSessionID)

	// For now, we'll return a simple list
	// In a full implementation, we'd iterate through m.core's listeners

	// TODO: Add method to listener.Manager to get all listeners
	// For now, return empty list
	listPayload := clientapi.ListenerListPayload{
		Listeners: []clientapi.ListenerStatusPayload{},
	}

	payloadBytes, _ := json.Marshal(listPayload)

	return clientapi.ServerResponse{
		Status:   clientapi.StatusSuccess,
		DataType: clientapi.DataTypeListenerList,
		Payload:  payloadBytes,
	}, nil
}

// StopListener handles operator requests to stop a listener.
func (m *OperatorListenerManager) StopListener(ctx context.Context, req clientapi.StopListenerPayload, operatorSessionID string) (clientapi.ServerResponse, error) {
	log.Printf("[LISTENER API] Operator %s requesting to stop listener %s",
		operatorSessionID, req.ListenerID)

	// Use the core manager to stop the listener
	if err := m.core.StopListener(req.ListenerID); err != nil {
		return clientapi.ServerResponse{
			Status: clientapi.StatusError,
			Error:  fmt.Sprintf("Failed to stop listener: %v", err),
		}, err
	}

	// Format success response
	confirmPayload := clientapi.ListenerStopConfirmationPayload{
		ListenerID: req.ListenerID,
		Message:    fmt.Sprintf("Listener %s stopped successfully by operator %s", req.ListenerID, operatorSessionID),
	}

	payloadBytes, _ := json.Marshal(confirmPayload)

	return clientapi.ServerResponse{
		Status:   clientapi.StatusSuccess,
		DataType: clientapi.DataTypeListenerStopConfirmation,
		Payload:  payloadBytes,
	}, nil
}

// Shutdown gracefully shuts down all listeners.
func (m *OperatorListenerManager) Shutdown(ctx context.Context) error {
	log.Println("[LISTENER API] Shutting down all listeners")
	m.core.StopAll()
	return nil
}

// mapProtocolToListenerType converts operator protocol names to internal listener types.
func (m *OperatorListenerManager) mapProtocolToListenerType(protocol string) (listener.ListenerType, error) {
	// Operators use the same names as agents (H1C, H1TLS, etc.)
	// We map these to internal listener types
	switch protocol {
	case "H1C":
		return listener.TypeHTTP1Clear, nil
	case "H1TLS":
		return listener.TypeHTTP1TLS, nil
	case "H2TLS":
		return listener.TypeHTTP2TLS, nil
	case "H3":
		return listener.TypeHTTP3, nil
	case "WS":
		return listener.TypeWebsocketClear, nil
	case "WSS":
		return listener.TypeWebsocketSecure, nil
	default:
		return "", fmt.Errorf("unknown protocol: %s", protocol)
	}
}
