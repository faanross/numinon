package clientapi

type EventType string

// Server-to-Client Event Types (for server-initiated messages)
const (
	EventTypeAgentCheckin EventType = "EVENT_AGENT_CHECKIN"
	EventTypeTaskResult   EventType = "EVENT_TASK_RESULT" // When an agent returns the result of a task
	EventTypeServerLog    EventType = "EVENT_SERVER_LOG"  // General logs/notifications from server
	EventTypeNewAgent     EventType = "EVENT_NEW_AGENT"   // When a new agent is first seen
)
