package orchestration

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
	"os"
	"path/filepath"
	"strings"
)

// DownloadOrchestrator handles the download command lifecycle
type DownloadOrchestrator struct {
	// BaseDir is the root directory for downloaded files
	BaseDir string
}

// NewDownloadOrchestrator creates a new download orchestrator
func NewDownloadOrchestrator(baseDir string) *DownloadOrchestrator {
	if baseDir == "" {
		baseDir = "./downloads"
	}
	return &DownloadOrchestrator{
		BaseDir: baseDir,
	}
}

// PrepareTask sets up download-specific metadata
func (d *DownloadOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first
	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid download arguments: %w", err)
	}

	// Parse arguments to get the source file name
	var args models.DownloadArgs
	if err := json.Unmarshal(task.Arguments, &args); err != nil {
		return fmt.Errorf("failed to parse download arguments: %w", err)
	}

	// Determine where to save the file on the server
	// Structure: BaseDir/agentID/timestamp_taskID_filename
	agentDir := filepath.Join(d.BaseDir, task.AgentID)

	// Extract just the filename from the source path
	sourceFileName := filepath.Base(args.SourceFilePath)
	if sourceFileName == "." || sourceFileName == "/" {
		sourceFileName = "unknown_file"
	}

	// Create a unique filename to avoid collisions
	// Format: taskID_originalname
	saveFileName := fmt.Sprintf("%s_%s", task.ID, sourceFileName)
	savePath := filepath.Join(agentDir, saveFileName)

	// Store metadata
	task.Metadata["server_save_path"] = savePath
	task.Metadata["source_file_path"] = args.SourceFilePath
	task.Metadata["source_file_name"] = sourceFileName

	log.Printf("|📋 ORCH DOWNLOAD| Prepared task %s: will save to %s", task.ID, savePath)

	return nil
}

// ProcessResult handles the downloaded file
func (d *DownloadOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {
	// Parse the agent's result
	var agentResult models.AgentTaskResult
	if err := json.Unmarshal(result, &agentResult); err != nil {
		return fmt.Errorf("failed to parse agent result: %w", err)
	}

	// Check if agent reported success
	if !strings.Contains(strings.ToLower(agentResult.Status), "success") {
		return fmt.Errorf("agent reported failure: %s - %s", agentResult.Status, agentResult.Error)
	}

	// Get the save path from metadata
	savePath, ok := task.Metadata["server_save_path"].(string)
	if !ok || savePath == "" {
		return fmt.Errorf("no save path found in task metadata")
	}

	// Decode the file content from base64
	if agentResult.Output == nil || len(agentResult.Output) == 0 {
		return fmt.Errorf("no file content in agent result")
	}

	// The Output field should contain base64-encoded file content
	var fileContentB64 string
	if err := json.Unmarshal(agentResult.Output, &fileContentB64); err != nil {
		// Maybe it's already a string, try direct conversion
		fileContentB64 = string(agentResult.Output)
		// Clean up if it has quotes
		fileContentB64 = strings.Trim(fileContentB64, "\"")
	}

	fileContent, err := base64.StdEncoding.DecodeString(fileContentB64)
	if err != nil {
		return fmt.Errorf("failed to decode file content from base64: %w", err)
	}

	// Verify hash if provided
	if agentResult.FileSha256 != "" {
		hasher := sha256.New()
		hasher.Write(fileContent)
		calculatedHash := hex.EncodeToString(hasher.Sum(nil))

		if calculatedHash != agentResult.FileSha256 {
			return fmt.Errorf("hash mismatch: expected %s, got %s",
				agentResult.FileSha256, calculatedHash)
		}
		log.Printf("|✅ ORCH DOWNLOAD| Hash verified for task %s: %s", task.ID, calculatedHash)
	}

	// Create directory if it doesn't exist
	saveDir := filepath.Dir(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", saveDir, err)
	}

	// Save the file
	if err := os.WriteFile(savePath, fileContent, 0644); err != nil {
		return fmt.Errorf("failed to save file to %s: %w", savePath, err)
	}

	// Update task metadata with results
	task.Metadata["file_saved"] = true
	task.Metadata["file_size"] = len(fileContent)
	task.Metadata["file_hash"] = agentResult.FileSha256

	log.Printf("|✅ ORCH DOWNLOAD| Successfully saved file for task %s: %s (%d bytes)",
		task.ID, savePath, len(fileContent))

	return nil
}

// ValidateArgs checks if download arguments are valid
func (d *DownloadOrchestrator) ValidateArgs(args json.RawMessage) error {
	var downloadArgs models.DownloadArgs
	if err := json.Unmarshal(args, &downloadArgs); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	if downloadArgs.SourceFilePath == "" {
		return fmt.Errorf("source_file_path is required")
	}

	// Could add more validation here (e.g., path traversal checks)

	return nil
}
