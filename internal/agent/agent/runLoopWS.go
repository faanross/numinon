package agent

import (
	"log"
)

// runWsLoop handles the persistent connection and message handling for WebSocket protocols.
func (a *Agent) runWsLoop() error {
	log.Printf("|AGENT LOOP WS|-> WebSocket loop starting with communicator type: %s.", a.communicator.Type())

	// essentially, it's in a infinite loop

	// it can receive messages, which will be instructions, it needs to hand this off to executeTask
	// This will of course be ServerTaskResponse

}
