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
	"numinon_shadow/internal/tracker"
	"strings"
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
	taskStore     taskmanager.ObservableTaskManager // Observer wraps taskmanager.TaskManger to give it notification abilities
	orchestrators *orchestration.Registry           // For task preparation/validation
	clientMgr     clientapi.ClientSessionManager    // To send results back to operators

	// Tracking mappings
	mu         sync.RWMutex
	taskOwners map[string]string // taskID -> operatorSessionID

	// Pusher allows for immediate push of tasks to WS-agents
	wsPusher     *WebSocketTaskPusher // For immediate WS delivery
	agentTracker *tracker.Tracker     // To check connection type
}

// NewBroker creates a new task broker with the necessary dependencies.
func NewBroker(
	taskStore taskmanager.ObservableTaskManager,
	orchestrators *orchestration.Registry,
	clientMgr clientapi.ClientSessionManager,
	wsPusher *WebSocketTaskPusher,
	agentTracker *tracker.Tracker,
) *Broker {
	broker := &Broker{
		taskStore:     taskStore,
		orchestrators: orchestrators,
		clientMgr:     clientMgr,
		wsPusher:      wsPusher,
		agentTracker:  agentTracker,
		taskOwners:    make(map[string]string),
	}

	// Subscribe to task events
	taskStore.Subscribe(broker)
	log.Println("[TASK BROKER] Subscribed to task store events")

	return broker
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

	//  Attempt immediate push for WebSocket agents
	if b.wsPusher != nil && b.agentTracker != nil {
		// Check if agent is WebSocket-connected
		agentInfo, exists := b.agentTracker.GetAgentInfo(agentID)
		if exists && agentInfo.ConnectionType == tracker.TypeWebSocket {
			// Agent is WebSocket-connected, try immediate push
			if err := b.wsPusher.PushTask(agentID, task); err != nil {
				log.Printf("[TASK BROKER] Failed to push task immediately: %v. Task queued for next check-in.", err)
				// Not a fatal error - task is queued and agent will get it eventually
			} else {
				// Mark as dispatched since we pushed it
				if err := b.taskStore.MarkDispatched(task.ID); err != nil {
					log.Printf("[TASK BROKER] Warning: Failed to mark pushed task as dispatched: %v", err)
				}
				log.Printf("[TASK BROKER] Task %s pushed immediately to WebSocket agent %s",
					task.ID, agentID)
			}
		} else {
			log.Printf("[TASK BROKER] Agent %s is not WebSocket-connected. Task queued for next check-in.", agentID)
		}
	}
	
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

	// The actual result storage and notification happens through the observer pattern
	// This method is here for any additional processing we might need

	// Verify the task exists and belongs to a known operator
	b.mu.RLock()
	operatorID, exists := b.taskOwners[result.TaskID]
	b.mu.RUnlock()

	if !exists {
		// This might be a task created before the broker was initialized
		// or a direct task not created through an operator
		log.Printf("[TASK BROKER] No operator tracking for task %s (might be a direct task)", result.TaskID)
		return nil // Not an error, just not operator-initiated
	}

	log.Printf("[TASK BROKER] Task %s belongs to operator %s", result.TaskID, operatorID)

	// The actual notification will happen through OnTaskCompleted when
	// the task store marks the task as complete

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

// OnTaskCompleted handles successful task completion.
// This is where we route results back to operators.
func (b *Broker) OnTaskCompleted(task *taskmanager.Task) {
	log.Printf("[TASK BROKER] Task %s completed for agent %s", task.ID, task.AgentID)

	// Look up which operator created this task
	b.mu.RLock()
	operatorID, exists := b.taskOwners[task.ID]
	b.mu.RUnlock()

	if !exists {
		log.Printf("[TASK BROKER] Warning: No operator found for completed task %s", task.ID)
		return
	}

	// Parse the raw result to get status info
	var agentResult models.AgentTaskResult
	if err := json.Unmarshal(task.Result, &agentResult); err != nil {
		log.Printf("[TASK BROKER] Failed to parse result for task %s: %v", task.ID, err)
		return
	}

	// Create operator-friendly result notification
	resultPayload := clientapi.TaskResultEventPayload{
		AgentID:        task.AgentID,
		TaskID:         task.ID,
		CommandType:    task.Command,
		ResultData:     task.Result, // Raw result data
		CommandSuccess: strings.Contains(agentResult.Status, "success"),
		ErrorMsg:       agentResult.Error,
	}

	payloadBytes, _ := json.Marshal(resultPayload)

	// Send to operator via client manager
	response := clientapi.ServerResponse{
		Status:   clientapi.StatusUpdate,
		DataType: clientapi.DataTypeCommandResult,
		Payload:  payloadBytes,
	}

	if err := b.clientMgr.SendToClient(operatorID, response); err != nil {
		log.Printf("[TASK BROKER] Failed to send result to operator %s: %v", operatorID, err)
	} else {
		log.Printf("[TASK BROKER] Result for task %s sent to operator %s", task.ID, operatorID)
	}
}

// OnTaskFailed handles task failure.
func (b *Broker) OnTaskFailed(task *taskmanager.Task, errorMsg string) {
	log.Printf("[TASK BROKER] Task %s failed: %s", task.ID, errorMsg)

	// Look up operator
	b.mu.RLock()
	operatorID, exists := b.taskOwners[task.ID]
	b.mu.RUnlock()

	if !exists {
		log.Printf("[TASK BROKER] Warning: No operator found for failed task %s", task.ID)
		return
	}

	// Create failure notification
	resultPayload := clientapi.TaskResultEventPayload{
		AgentID:        task.AgentID,
		TaskID:         task.ID,
		CommandType:    task.Command,
		CommandSuccess: false,
		ErrorMsg:       errorMsg,
	}

	payloadBytes, _ := json.Marshal(resultPayload)

	response := clientapi.ServerResponse{
		Status:   clientapi.StatusUpdate,
		DataType: clientapi.DataTypeCommandResult,
		Payload:  payloadBytes,
	}

	if err := b.clientMgr.SendToClient(operatorID, response); err != nil {
		log.Printf("[TASK BROKER] Failed to send failure notification to operator %s: %v", operatorID, err)
	}
}

// OnTaskDispatched handles task dispatch notification.
func (b *Broker) OnTaskDispatched(task *taskmanager.Task) {
	log.Printf("[TASK BROKER] Task %s dispatched to agent %s", task.ID, task.AgentID)
	// For now, we just log.
	// TODO Could send a status update to operator if desired.
}
