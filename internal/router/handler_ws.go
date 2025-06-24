package router

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func WSHandler(w http.ResponseWriter, r *http.Request) {

	// As with all websocket handlers first thing we do is upgrade connection
	// We then of course also need to create this struct (see below this function)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("|❗ERR| WS Upgrade Failed: %v", err)
		return
	}
	defer conn.Close()

	// At this point now we are upgraded - ie from here on fwd WE HAVE AN ACTUAL WEBSOCKET CONNECTION!
	log.Printf("|🔌 WS| Client connected from %s", conn.RemoteAddr())

	// This is the classic WS pattern - an infinite for loop
	// This will be called from both client and server side
	// This "locks" them into the bidirectional push state

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// Handle errors, including client disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected errors
				log.Printf("|❗ERR| WS Read Error: %v", err)
			} else {
				// Log expected closures (client disconnected)
				log.Printf("|🔌 WS| Client %s disconnected.", conn.RemoteAddr())
			}
			break // Exit the loop on error or disconnection
		}

		// Display received message
		log.Printf("|💬 WS| Received from %s: %s", conn.RemoteAddr(), string(message))

		// Echo message back to client
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("|❗ERR| WS Write Error: %v", err)
			break
		}
	}

}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Right now we are indiscriminately upgrading WS connection
	// Later, we def should discriminate!
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
