package taskmanager

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// MemoryTaskStore implements TaskManager using in-memory storage
type MemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
	rng   *rand.Rand
}

// NewMemoryTaskStore creates a new in-memory task store
func NewMemoryTaskStore() *MemoryTaskStore {
	return &MemoryTaskStore{
		tasks: make(map[string]*Task),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateTask creates and stores a new task
func (m *MemoryTaskStore) CreateTask(agentID, command string, args json.RawMessage) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique task ID
	taskID := m.generateTaskID()

	// Create new task
	task := NewTask(taskID, agentID, command, args)

	// Store in map
	m.tasks[taskID] = task

	return task, nil
}

// GetTask retrieves a task by ID
func (m *MemoryTaskStore) GetTask(taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// UpdateTask updates an existing task
func (m *MemoryTaskStore) UpdateTask(task *Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}

	m.tasks[task.ID] = task
	return nil
}

// MarkDispatched updates task status when sent to agent
func (m *MemoryTaskStore) MarkDispatched(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	// Validate state transition
	if task.Status != StatusPending {
		return fmt.Errorf("%w: cannot dispatch task in status %s", ErrInvalidState, task.Status)
	}

	task.MarkDispatched()
	return nil
}

// StoreResult stores result and marks task completed
func (m *MemoryTaskStore) StoreResult(taskID string, result json.RawMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	// Only dispatched tasks can be completed
	if task.Status != StatusDispatched {
		return fmt.Errorf("%w: cannot complete task in status %s", ErrInvalidState, task.Status)
	}

	task.MarkCompleted(result)
	return nil
}

// MarkFailed marks a task as failed
func (m *MemoryTaskStore) MarkFailed(taskID string, errorMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	// Can fail from dispatched or pending state
	if task.IsTerminal() {
		return fmt.Errorf("%w: cannot fail task already in terminal state %s", ErrInvalidState, task.Status)
	}

	task.MarkFailed(errorMsg)
	return nil
}

// GetAgentTasks returns all tasks for a specific agent
func (m *MemoryTaskStore) GetAgentTasks(agentID string) ([]*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agentTasks []*Task
	for _, task := range m.tasks {
		if task.AgentID == agentID {
			agentTasks = append(agentTasks, task)
		}
	}

	return agentTasks, nil
}

// GetPendingTasks returns tasks that haven't been dispatched yet
func (m *MemoryTaskStore) GetPendingTasks() ([]*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pendingTasks []*Task
	for _, task := range m.tasks {
		if task.Status == StatusPending {
			pendingTasks = append(pendingTasks, task)
		}
	}

	return pendingTasks, nil
}

// generateTaskID creates a unique task identifier
func (m *MemoryTaskStore) generateTaskID() string {
	// Simple format: task_XXXXXX
	return fmt.Sprintf("task_%06d", m.rng.Intn(1000000))
}

// Cleanup removes completed tasks older than the specified duration
// This is not yet implemented (TODO)
func (m *MemoryTaskStore) Cleanup(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, task := range m.tasks {
		if task.IsTerminal() && task.CompletedAt != nil && task.CompletedAt.Before(cutoff) {
			delete(m.tasks, id)
			removed++
		}
	}

	return removed
}

// Stats returns basic statistics about stored tasks
// This is not yet implemented (TODO)
func (m *MemoryTaskStore) Stats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total":      len(m.tasks),
		"pending":    0,
		"dispatched": 0,
		"completed":  0,
		"failed":     0,
		"timeout":    0,
	}

	for _, task := range m.tasks {
		stats[string(task.Status)]++
	}

	return stats
}
