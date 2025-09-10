package clientapi

type ActionType string

// Client-to-Server Actions
const (
	// Listener Management
	ActionCreateListener ActionType = "CREATE_LISTENER"
	ActionListListeners  ActionType = "LIST_LISTENERS"
	ActionStopListener   ActionType = "STOP_LISTENER"

	// Agent Command Instructions
	ActionTaskAgentUploadFile       ActionType = "TASK_AGENT_UPLOAD_FILE"
	ActionTaskAgentDownloadFile     ActionType = "TASK_AGENT_DOWNLOAD_FILE"
	ActionTaskAgentRunCmd           ActionType = "TASK_AGENT_RUN_CMD"
	ActionTaskAgentExecuteShellcode ActionType = "TASK_AGENT_EXECUTE_SHELLCODE"
	ActionTaskAgentEnumerateProcs   ActionType = "TASK_AGENT_ENUMERATE_PROCESSES"
	ActionTaskAgentMorph            ActionType = "TASK_AGENT_MORPH"
	ActionTaskAgentHop              ActionType = "TASK_AGENT_HOP"

	// Agent & Server Management
	ActionListAgents      ActionType = "LIST_AGENTS"
	ActionGenerateAgent   ActionType = "GENERATE_AGENT_BINARY"
	ActionGetServerStatus ActionType = "GET_SERVER_STATUS"
)
