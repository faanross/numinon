package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"os"
)

func DoDownload(args models.DownloadArgs) (models.DownloadResult, error) {

	// Display command arguments to terminal (TODO convert to slog)
	fmt.Println("|✅ DOWNLOAD DOER| The DOWNLOAD command has been executed.")
	fmt.Printf("|📋 DOWNLOAD DETAILS| SoureFilePath: '%s'\n", args.SourceFilePath)

// validate essential arguments
	if args.SourceFilePath == "" {
		return DownloadResult{}, fmt.Errorf("validation: SourceFilePath cannot be empty")
	}






}

// THIS IS THE REFERENCE HERE BELOW





	// 1. Read the file from agent's disk
	rawFileBytes, err := os.ReadFile(args.SourceFilePath)
	if err != nil {
		// Differentiate common errors
		if os.IsNotExist(err) {
			msg := fmt.Sprintf("File not found at '%s'", args.SourceFilePath)
			log.Printf("|❗ERR CMD DOWNLOAD_EXEC| %s", msg)
			return DownloadResult{SourcePath: args.SourceFilePath, Message: msg}, fmt.Errorf(msg) // Return specific error
		}
		if os.IsPermission(err) {
			msg := fmt.Sprintf("Permission denied reading file '%s'", args.SourceFilePath)
			log.Printf("|❗ERR CMD DOWNLOAD_EXEC| %s", msg)
			return DownloadResult{SourcePath: args.SourceFilePath, Message: msg}, fmt.Errorf(msg)
		}
		// General read error
		msg := fmt.Sprintf("Failed to read file '%s'", args.SourceFilePath)
		log.Printf("|❗ERR CMD DOWNLOAD_EXEC| %s: %v", msg, err)
		return DownloadResult{SourcePath: args.SourceFilePath, Message: msg}, fmt.Errorf("%s: %w", msg, err)
	}
	log.Printf("|CMD DOWNLOAD_EXEC| Successfully read %d bytes from '%s'.", len(rawFileBytes), args.SourceFilePath)

	// 2. Calculate SHA256 hash of the raw file bytes
	hasher := sha256.New()
	hasher.Write(rawFileBytes)
	fileSha256 := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("|CMD DOWNLOAD_EXEC| Calculated SHA256 for '%s': %s", args.SourceFilePath, fileSha256)

	successMsg := fmt.Sprintf("Successfully read and hashed file '%s' (%d bytes). SHA256: %s",
		args.SourceFilePath, len(rawFileBytes), fileSha256)

	return DownloadResult{
		RawFileBytes: rawFileBytes,
		SourcePath:   args.SourceFilePath,
		FileSha256:   fileSha256,
		Message:      successMsg,
	}, nil // Success
}
