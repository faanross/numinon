package clientapi

import (
	"context"
	"github.com/faanross/numinon/internal/models"
	"github.com/gorilla/websocket"
)

// ClientSession represents an active operator client connection.
type ClientSession interface {
	ID() string                         // Returns the unique session ID.
	Send(response ServerResponse) error // Sends a message to this specific client.
	Close() error                       // Closes the client session and its underlying connection.
	RemoteAddr() string                 // Returns the remote address of the client.
}

// ClientSessionManager defines the contract for managing active operator client sessions.
type ClientSessionManager interface {
	Register(conn *websocket.Conn) (sessionID string, err error)  // Called when a new client session is established.
	Unregister(sessionID string) error                            // Called when a client session ends (disconnects or error).
	DispatchRequest(sessionID string, req ClientRequest)          // Handles an incoming ClientRequest from a specific client session.
	SendToClient(sessionID string, response ServerResponse) error // Allows server to send a ServerResponse to a specific client session
	Broadcast(response ServerResponse)                            // Sends a message to all currently active/registered client sessions.
}

// TODO this will replace the listener manager in pkg listener
// This one adds an extra dimension - connection client <-> server <-> agent
// TODO Once this is fully implemented, transition from old one, and remove it

// ListenerManager defines operations for managing listeners.
type ListenerManager interface {
	CreateListener(ctx context.Context, req CreateListenerPayload, operatorSessionID string) (ServerResponse, error)
	StopListener(ctx context.Context, req StopListenerPayload, operatorSessionID string) (ServerResponse, error)
	ListListeners(ctx context.Context, operatorSessionID string) (ServerResponse, error)
	Shutdown(ctx context.Context) error
}

// TaskBroker acts as a bridge between operator requests and the core task management system.
// It translates operator commands into tasks, delegates to the actual TaskManager,
// and ensures results are routed back to the correct operator.
// This is NOT a task storage system - it's a coordinator that uses the existing TaskManager from pkg taskmanager

type TaskBroker interface { // <- Changed from TaskManager
	// QueueAgentTask takes a request from an operator and queues it for the specified agent.
	QueueAgentTask(ctx context.Context, req ClientRequest, operatorSessionID string) (ServerResponse, error)

	// ProcessAgentResult takes a completed task result from an agent and forwards the result to the operator.
	ProcessAgentResult(actualResult models.AgentTaskResult) error

	// GetTaskOwner returns the operator session ID that created a specific task
	GetTaskOwner(taskID string) (string, error)
}

// AgentStateManager defines operations for managing known agent state from the API layer.
// Note: This is the public interface for the API, not the internal manager itself.
type AgentStateManager interface {
	// ListAgents returns a summary of all known agents.
	ListAgents(ctx context.Context, operatorSessionID string) (ServerResponse, error)

	// GetAgentDetails returns detailed information for a single agent.
	GetAgentDetails(ctx context.Context, agentID string, operatorSessionID string) (ServerResponse, error)

	// InitiateHop signals an agent to HOP, allowing for stateful reporting
	InitiateHop(agentID, taskID, listenerID string, cancelFunc context.CancelFunc)

	// ClearHopState removes the pending hop state from an agent following success/failure
	ClearHopState(agentID string)
}
