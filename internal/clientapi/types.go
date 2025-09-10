package clientapi

type DataType string

// Server-to-Client Data Types (for ServerResponse.DataType)
const (
	DataTypeListenerStatus           DataType = "LISTENER_STATUS"
	DataTypeListenerList             DataType = "LISTENER_LIST"
	DataTypeTaskQueuedConfirmation   DataType = "TASK_QUEUED_CONFIRMATION"
	DataTypeAgentList                DataType = "AGENT_LIST"
	DataTypeAgentDetails             DataType = "AGENT_DETAILS"
	DataTypeAgentError               DataType = "AGENT_ERROR"    // If an agent reports an error performing a task
	DataTypeCommandResult            DataType = "COMMAND_RESULT" // General structure for results from agent commands
	DataTypeServerLogEntry           DataType = "SERVER_LOG_ENTRY"
	DataTypeNewAgentNotification     DataType = "NEW_AGENT_NOTIFICATION"
	DataTypeGeneratedAgentConfig     DataType = "GENERATED_AGENT_CONFIG" // For ActionGenerateAgent response
	DataTypeError                    DataType = "ERROR_DETAILS"          // For a more structured error in Payload
	DataTypeListenerStopConfirmation string   = "LISTENER_STOP_CONFIRMATION"
	DataTypeErrorDetails             string   = "ERROR_DETAILS" // Generic for structured error payloads
)
