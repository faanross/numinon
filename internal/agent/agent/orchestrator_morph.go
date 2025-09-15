package agent

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/models"
	"log"
	"strings"
	"time"
)

// orchestrateMorph is the orchestrator for the MORPH command.
func (a *Agent) orchestrateMorph(task models.ServerTaskResponse) models.AgentTaskResult {
	var args models.MorphArgs

	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal MorphArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))

		log.Printf("|❗ERR MORPH ORCHESTRATOR| %s", errMsg)

		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureInvalidArgs,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ ENUMERATE ORCHESTRATOR| Task ID: %s. Received MorphArgs: %+v", task.TaskID, args)

	// this allows us to keep track of what individual parameters failed/succeed during morph
	var updateMessages []string
	configChanged := false    // Tracks if any *supported* config was actually changed
	validationFailed := false // Tracks if any *supported* and *attempted* parameter failed validation

	// Check if a new DELAY value has been issued
	if args.NewDelay != nil {
		log.Printf("|⚙️ MORPH ACTION| Attempting to MORPH Delay to: '%s'", *args.NewDelay)

		// parse and convert string input to time.Duration
		newSleepDuration, parseErr := time.ParseDuration(*args.NewDelay)

		if parseErr != nil {
			errMsg := fmt.Sprintf("BaseSleep update failed: invalid duration format '%s'. Error: %v", *args.NewDelay, parseErr)
			log.Printf("|❗WARNING MORPH ORCHESTRATOR| %s", errMsg)
			updateMessages = append(updateMessages, errMsg)
			validationFailed = true
		} else if newSleepDuration <= 0 {
			errMsg := fmt.Sprintf("BaseSleep update failed: duration '%v' must be positive.", newSleepDuration)
			log.Printf("|❗WARNING MORPH ORCHESTRATOR| %s", errMsg)
			updateMessages = append(updateMessages, errMsg)
			validationFailed = true
		} else {
			a.config.Delay = newSleepDuration
			msg := fmt.Sprintf("BaseSleep successfully updated to: %v", newSleepDuration)
			log.Printf("|✅ ENUMERATE ORCHESTRATOR| %s", msg)
			updateMessages = append(updateMessages, msg)
			configChanged = true
		}
	}

	// Check if a new JITTER value has been issued
	if args.NewJitter != nil {
		log.Printf("|⚙️ MORPH ACTION| Attempting to morph Jitter to: %f", *args.NewJitter)
		newJitter := *args.NewJitter

		if newJitter < 0.0 || newJitter > 1.0 {
			errMsg := fmt.Sprintf("Jitter update failed: value %f is out of acceptable range [0.0 - 1.0].", newJitter)
			log.Printf("|❗WARNING MORPH ORCHESTRATOR| %s", errMsg)
			updateMessages = append(updateMessages, errMsg)
			validationFailed = true
		} else {
			a.config.Jitter = newJitter
			msg := fmt.Sprintf("Jitter successfully updated to: %f", newJitter)
			log.Printf("|✅ ENUMERATE ORCHESTRATOR| %s", msg)
			updateMessages = append(updateMessages, msg)
			configChanged = true
		}
	}

	// --- Determine Final Status ---
	finalStatus := models.StatusFailureMorphNoValidChanges // Default if nothing changed or only ignored params were sent

	if configChanged {
		if validationFailed {
			finalStatus = models.StatusSuccessMorphPartial
		} else { // All attempted (and supported) changes were successful
			finalStatus = models.StatusSuccessMorphApplied
		}
	} else if validationFailed {
		// No supported config was changed, but there was an attempt to change a supported param (BaseSleep/Jitter) which failed validation.
		finalStatus = models.StatusFailureInvalidArgs
	}
	// If only "not supported" messages are present in updateMessages, and no valid changes made to BaseSleep/Jitter,
	// and no validation failures for BaseSleep/Jitter, then finalStatus remains StatusFailureMorphNoValidChanges.

	log.Printf("|✅ ENUMERATE ORCHESTRATOR| Finalizing morph task. Status: %s. Output messages: %s", finalStatus, strings.Join(updateMessages, "; "))

	finalOutput := strings.Join(updateMessages, "; ")
	outputJSON, err := json.Marshal(finalOutput)
	if err != nil {
		// This should never happen for a string, but just in case
		log.Printf("|❗ERR MORPH ORCHESTRATOR| Failed to marshal output: %v", err)
		outputJSON = []byte("null") // Fallback to valid JSON
	}

	return models.AgentTaskResult{
		TaskID: task.TaskID,
		Status: finalStatus,
		Output: outputJSON,
		Error:  "", // Specific validation errors are part of the Output messages.
	}
}
