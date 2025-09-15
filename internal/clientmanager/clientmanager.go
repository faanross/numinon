package clientmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/clientapi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 24 * time.Hour
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2097152
)

// clientSession represents a single connected client.
// It acts as a bridge between the WebSocket connection and the Client Session manager.
type clientSession struct {
	manager *clientSessionManagerImpl // A reference to the central manager.

	conn *websocket.Conn // The WebSocket connection for this client session.

	// A buffered channel of outbound messages. Messages are placed here by the manager
	// and sent to the WebSocket connection by the writePump goroutine.
	// This decouples message sending from the business logic.
	send chan []byte

	// A unique ID for this session, assigned upon registration.
	sessionID string
}

// clientSessionManagerImpl is the central hub for managing all client sessions.
// It maintains the registry of active sessions and routes messages.
type clientSessionManagerImpl struct {
	sessions   map[string]*clientSession
	register   chan *clientSession
	unregister chan *clientSession
	broadcast  chan []byte

	// --- Dependencies on other managers ---
	listenerManager clientapi.ListenerManager
	taskBroker      clientapi.TaskBroker
	agentStateMgr   clientapi.AgentStateManager
}

// New creates a new clientSessionManagerImpl and starts its main run loop as a goroutine.
// Remember -> There is one goroutine for manager as a whole (implicitly thread-safe) + 2 per session (R + W pumps)
func New(lm clientapi.ListenerManager, tb clientapi.TaskBroker, asm clientapi.AgentStateManager) *clientSessionManagerImpl {
	csm := &clientSessionManagerImpl{
		sessions:        make(map[string]*clientSession),
		register:        make(chan *clientSession),
		unregister:      make(chan *clientSession),
		broadcast:       make(chan []byte),
		listenerManager: lm,
		taskBroker:      tb,
		agentStateMgr:   asm,
	}
	// Launch the central hub in the background.
	go csm.run()
	return csm
}

// run is the main event loop for the session manager, which is started as a goroutine.
// It listens on its channels and safely modifies the sessions map.
func (m *clientSessionManagerImpl) run() {
	for {
		select {
		case session := <-m.register:
			// A new client has connected. Add them to our map.
			m.sessions[session.sessionID] = session
			log.Printf("New CLIENT session %s registered from %s", session.sessionID, session.conn.RemoteAddr())

		case session := <-m.unregister:
			// A client has disconnected.
			if _, ok := m.sessions[session.sessionID]; ok {
				delete(m.sessions, session.sessionID)
				close(session.send) // Close the send channel to signal the writePump to exit.
				log.Printf("CLIENT session %s unregistered", session.sessionID)
			}

		case message := <-m.broadcast:
			// A message needs to be sent to everyone.
			for sessionID, session := range m.sessions {
				select {
				case session.send <- message:
				default:
					// If the send channel is full, the client is lagging.
					// We close the channel and remove the session.
					close(session.send)
					delete(m.sessions, sessionID)
				}
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the manager's dispatcher.
func (s *clientSession) readPump() {
	defer func() {
		s.manager.unregister <- s
		s.conn.Close()
	}()
	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error { s.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var req clientapi.ClientRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Error unmarshalling operator request: %v", err)
			// TODO: Send a structured error back to the client.
			continue
		}

		// Request has been received, use `s.manager` reference to hand off the work to the central hub.
		s.manager.DispatchRequest(s.sessionID, req)
	}
}

// writePump pumps messages from the send channel to the websocket connection.
func (s *clientSession) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()
	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (m *clientSessionManagerImpl) Register(conn *websocket.Conn) (string, error) {
	sessionID := "op_session_" + uuid.New().String()
	session := &clientSession{
		manager:   m,
		conn:      conn,
		send:      make(chan []byte, 256),
		sessionID: sessionID,
	}
	m.register <- session
	go session.writePump()
	go session.readPump()
	return sessionID, nil
}

func (m *clientSessionManagerImpl) Unregister(sessionID string) error {
	if session, ok := m.sessions[sessionID]; ok {
		m.unregister <- session
	}
	return nil
}

func (m *clientSessionManagerImpl) DispatchRequest(sessionID string, req clientapi.ClientRequest) {
	var response clientapi.ServerResponse

	var err error

	ctx := context.Background() // Base context for manager calls

	switch req.Action {
	case clientapi.ActionCreateListener:
		var payload clientapi.CreateListenerPayload
		if err = json.Unmarshal(req.Payload, &payload); err == nil {
			response, err = m.listenerManager.CreateListener(ctx, payload, sessionID)
		}
	case clientapi.ActionListListeners:
		response, err = m.listenerManager.ListListeners(ctx, sessionID)
	case clientapi.ActionStopListener:
		var payload clientapi.StopListenerPayload
		if err = json.Unmarshal(req.Payload, &payload); err == nil {
			response, err = m.listenerManager.StopListener(ctx, payload, sessionID)
		}

	case clientapi.ActionTaskAgentRunCmd,
		clientapi.ActionTaskAgentUploadFile,
		clientapi.ActionTaskAgentDownloadFile,
		clientapi.ActionTaskAgentExecuteShellcode,
		clientapi.ActionTaskAgentEnumerateProcs,
		clientapi.ActionTaskAgentMorph,
		clientapi.ActionTaskAgentHop,
		clientapi.ActionTaskAgentReportConfig:

		var payload clientapi.TaskAgentPayload
		if err = json.Unmarshal(req.Payload, &payload); err == nil {
			response, err = m.taskBroker.QueueAgentTask(ctx, req, sessionID)
		}

	case clientapi.ActionListAgents:
		response, err = m.agentStateMgr.ListAgents(ctx, sessionID)
	case clientapi.ActionGetAgentDetails:
		var payload clientapi.GetAgentDetailsPayload
		if err = json.Unmarshal(req.Payload, &payload); err == nil {
			response, err = m.agentStateMgr.GetAgentDetails(ctx, payload.AgentID, sessionID)
		}
	default:
		err = fmt.Errorf("unknown action: %s", req.Action)
	}

	// This error handling block primarily catches issues from payload unmarshalling
	// or catastrophic errors from the managers.
	if err != nil {
		log.Printf("Error dispatching request for session %s, action %s: %v", sessionID, req.Action, err)
		errPayload := clientapi.ErrorDetailsPayload{Detail: err.Error()}
		payloadBytes, _ := json.Marshal(errPayload)
		response = clientapi.ServerResponse{
			RequestID: req.RequestID,
			Status:    clientapi.StatusError,
			Error:     err.Error(),
			Payload:   payloadBytes,
		}
	} else {
		// The request ID from the original request must be attached to the response
		// so the client can correlate them.
		response.RequestID = req.RequestID
	}

	m.SendToClient(sessionID, response)
}

func (m *clientSessionManagerImpl) SendToClient(sessionID string, response clientapi.ServerResponse) error {
	if session, ok := m.sessions[sessionID]; ok {
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("failed to marshal response for session %s: %w", sessionID, err)
		}
		select {
		case session.send <- responseBytes:
		default:
			return fmt.Errorf("send channel full for session %s, client may be lagging or disconnected", sessionID)
		}
		return nil
	}
	return fmt.Errorf("session not found: %s", sessionID)
}

func (m *clientSessionManagerImpl) Broadcast(response clientapi.ServerResponse) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}
	m.broadcast <- responseBytes
}

var _ clientapi.ClientSessionManager = (*clientSessionManagerImpl)(nil)
