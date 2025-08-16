package models

// ------------------------------------------------------------------
// | All the command-specific RESULTS structs, i.e. AGENT -> SERVER |
// ------------------------------------------------------------------

// ProcessInfo holds information about a single running process (part of Enumerate result).
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

// UploadResult holds the specific outcomes of the upload operation.
type UploadResult struct {
	FilePath     string // The path where the file was written
	BytesWritten int64
	ActualSha256 string
	Message      string
	HashMatched  bool
}

// DownloadResult holds the specific outcomes of the upload operation.
type DownloadResult struct {
	RawFileBytes []byte // The actual content of the file read from disk
	SourcePath   string // The path from which the file was read on the agent
	FileSha256   string // SHA256 hash of RawFileBytes
	Message      string // Any informational message for logging/output
}
