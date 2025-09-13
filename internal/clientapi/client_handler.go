package clientapi

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// CSM is a package-level variable that will hold the global ClientSessionManager instance
var CSM ClientSessionManager

// TM is a package-level variable that will hold the global TaskManager instance
var TM TaskManager

// upgrader updates HTTP -> WS
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ClientWSHandler handles WebSocket connections from our client(s)
func ClientWSHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("|👂 CLIENT_API|-> Received request at /client endpoint. Attempting WebSocket upgrade...")

	if CSM == nil {
		log.Println("|❗ERR CLIENT_API|-> CRITICAL: ClientSessionManager (CSM) is not initialized.")
		http.Error(w, "Server configuration error: ClientSessionManager not available.", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("|❗ERR CLIENT_API|-> Failed to upgrade WebSocket connection for operator: %v", err)
		return
	}
	log.Printf("|🚀 CLIENT_API|-> Operator client WebSocket connection established from %s", conn.RemoteAddr())

	sessionID, err := CSM.Register(conn)
	if err != nil {
		log.Printf("|❗ERR CLIENT_API|-> Failed to register client session for %s: %v", conn.RemoteAddr(), err)
		conn.Close() // Close the raw connection as session registration failed
		return
	}

	log.Printf("|👂 CLIENT_API|-> ClientWSHandler completed upgrade and launched session read loop for session %s (%s).", sessionID, conn.RemoteAddr())
}

//
//// sendErrorResponse is a helper to send a structured error back to the client directly from the read loop
//func sendErrorResponse(conn *websocket.Conn, sessionIDForLog string, requestID string, errorMessage string) {
//	type ErrorDetailPayload struct {
//		Detail string `json:"detail"`
//	}
//	errorDetail := ErrorDetailPayload{Detail: errorMessage}
//	payloadBytes, marshalErr := json.Marshal(errorDetail)
//	if marshalErr != nil {
//		log.Printf("|❗ERR CLIENT_API|-> (sendErrorResponse) Failed to marshal error detail payload for session %s: %v. Using raw string.", sessionIDForLog, marshalErr)
//		// Fallback to a simpler raw JSON message if marshalling the struct fails
//		payloadBytes = json.RawMessage(fmt.Sprintf(`{"detail": "Error creating structured error: %s"}`, errorMessage))
//	}
//
//	errResponse := ServerResponse{
//		RequestID: requestID, // May be empty if original request_id couldn't be parsed
//		Status:    StatusError,
//		Error:     errorMessage, // Main error message
//		DataType:  DataTypeErrorDetails,
//		Payload:   json.RawMessage(payloadBytes),
//	}
//
//	responseBytes, err := json.Marshal(errResponse)
//	if err != nil {
//		log.Printf("|❗ERR CLIENT_API|-> (sendErrorResponse) Failed to marshal ServerResponse for session %s: %v", sessionIDForLog, err)
//		return
//	}
//
//	if writeErr := conn.WriteMessage(websocket.TextMessage, responseBytes); writeErr != nil {
//		log.Printf("|❗ERR CLIENT_API|-> (sendErrorResponse) Failed to send error response to client %s (session %s): %v", conn.RemoteAddr(), sessionIDForLog, writeErr)
//	}
//}
