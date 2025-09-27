// Package frontend also contains the translation logic for converting internal
// backend models into their frontend-safe DTO counterparts.
package frontend

import (
	"numinon-ui/internal/models"
	"time"
)

// ToConnectionStatusDTO converts an internal ConnectionStatus model to its DTO.
func ToConnectionStatusDTO(status models.ConnectionStatus) ConnectionStatusDTO {
	return ConnectionStatusDTO{
		Connected: status.Connected,
		ServerURL: status.ServerURL,
		LastPing:  status.LastPing.Format(time.RFC3339),
		Latency:   status.Latency,
		Error:     status.Error,
	}
}

// ToServerMessageDTO converts an internal ServerMessage model to its DTO.
func ToServerMessageDTO(msg models.ServerMessage) ServerMessageDTO {
	return ServerMessageDTO{
		Type:      msg.Type,
		Timestamp: msg.Timestamp.Format(time.RFC3339),
		Payload:   msg.Payload,
	}
}

// ToAgentDTO converts an internal Agent model to its DTO representation.
func ToAgentDTO(agent models.Agent) AgentDTO {
	return AgentDTO{
		ID:        agent.ID,
		Hostname:  agent.Hostname,
		OS:        agent.OS,
		Status:    agent.Status,
		LastSeen:  agent.LastSeen.Format(time.RFC3339),
		IPAddress: agent.IPAddress,
	}
}

// ToAgentDTOs converts a slice of internal Agent models to their DTOs.
func ToAgentDTOs(agents []models.Agent) []AgentDTO {
	dtos := make([]AgentDTO, len(agents))
	for i, agent := range agents {
		dtos[i] = ToAgentDTO(agent)
	}
	return dtos
}

// ToCommandRequestDTO converts a CommandRequest to its DTO.
func ToCommandRequestDTO(req models.CommandRequest) CommandRequestDTO {
	return CommandRequestDTO{
		AgentID:   req.AgentID,
		Command:   req.Command,
		Arguments: req.Arguments,
	}
}

// ToCommandResponseDTO converts a CommandResponse to its DTO.
func ToCommandResponseDTO(resp models.CommandResponse) CommandResponseDTO {
	return CommandResponseDTO{
		Success: resp.Success,
		Output:  resp.Output,
		Error:   resp.Error,
	}
}
