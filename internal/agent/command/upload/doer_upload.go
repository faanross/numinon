package upload

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/faanross/numinon/internal/models"
	"log"
	"os"
	"path/filepath"
)

// DoUpload executes the file upload logic.
func DoUpload(args models.UploadArgs) (models.UploadResult, error) {

	// Display command arguments to terminal (TODO convert to slog)
	fmt.Println("|✅ UPLOAD DOER| The UPLOAD command has been executed.")
	fmt.Printf("|📋 UPLOAD DETAILS| TargetDir: '%s', TargetFilename: '%s', Overwrite: %t, ExpectedSHA256: %s, ContentLen(b64): %d\n",
		args.TargetDirectory, args.TargetFilename, args.OverwriteIfExists, args.ExpectedSha256, len(args.FileContentBase64))

	// Validate essential arguments
	if args.TargetDirectory == "" {
		return models.UploadResult{}, fmt.Errorf("TargetDirectory cannot be empty")
	}
	if args.TargetFilename == "" {
		return models.UploadResult{}, fmt.Errorf("TargetFilename cannot be empty")
	}
	if args.FileContentBase64 == "" {
		return models.UploadResult{}, fmt.Errorf("FileContentBase64 cannot be empty")
	}
	if args.ExpectedSha256 == "" {
		return models.UploadResult{}, fmt.Errorf("ExpectedSha256 cannot be empty")
	}

	// Decode File Content
	rawFileBytes, err := base64.StdEncoding.DecodeString(args.FileContentBase64)
	if err != nil {
		return models.UploadResult{}, fmt.Errorf("Failed to base64 decode file content.")
	}
	log.Printf("|⚙️ UPLOAD ACTION| Decoded file content: %d bytes.", len(rawFileBytes))

	// Construct Full Destination Path

	// Basic sanitization: ensure filename doesn't try to escape the directory.
	// filepath.Clean will help, but more robust sandboxing might be needed eventually
	cleanFilename := filepath.Base(args.TargetFilename) // Use only the filename part
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "" {
		cleanFilename = "uploaded_file_unspecified_name" // Default if filename is problematic
		log.Printf("|CMD UPLOAD_EXEC| Original TargetFilename was problematic, using default: %s", cleanFilename)
	}

	// Ensure TargetDirectory exists
	// TODO if not create it, for now, assume it must exist or OS handles error
	// TODO for OPSEC sanitation, ensure TargetDirectory is within an allowed base path
	destinationPath := filepath.Join(args.TargetDirectory, cleanFilename)
	log.Printf("|⚙️ UPLOAD ACTION| Full destination path: %s\n", destinationPath)

	// Ensure target directory exists. If not, attempt to create it.
	// This makes the command more robust if the operator specifies a new subdir.
	dir := filepath.Dir(destinationPath)
	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		log.Printf("|⚙️ UPLOAD ACTION| Target directory '%s' does not exist. Attempting to create.", dir)
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil { // 0755: rwxr-xr-x
			return models.UploadResult{},
				fmt.Errorf("Failed to create target directory '%s'.", dir)
		}
		log.Printf("|⚙️ UPLOAD ACTION| Successfully created target directory '%s'.", dir)
	}
	// Check for Existing File & Overwrite Logic
	_, statErr := os.Stat(destinationPath)
	fileExists := !os.IsNotExist(statErr)

	if fileExists && !args.OverwriteIfExists {
		msg := fmt.Sprintf("File '%s' already exists and overwrite is not permitted.", destinationPath)
		log.Printf("|⚙️ UPLOAD ACTION| %s", msg)
		return models.UploadResult{}, fmt.Errorf("file '%s' already exists and overwrite is not permitted", destinationPath)
	}
	if fileExists && args.OverwriteIfExists {
		log.Printf("|⚙️ UPLOAD ACTION| File '%s' exists and will be overwritten as per OverwriteIfExists=true.", destinationPath)
	}

	// Write File to Disk
	// Using 0600 for more restrictive permissions initially.
	err = os.WriteFile(destinationPath, rawFileBytes, 0600)
	if err != nil {
		// TODO: More granular error detection (permissions, disk full) would be ideal here.
		// For now, any WriteFile error is treated generally.
		msg := fmt.Sprintf("Failed to write file to '%s'.", destinationPath)
		log.Printf("|⚙️ UPLOAD ACTION| %s: %v", msg, err)
		return models.UploadResult{}, fmt.Errorf("Failed to write file to '%s'.", destinationPath)
	}
	bytesWritten := int64(len(rawFileBytes))
	log.Printf("|⚙️ UPLOAD ACTION| Successfully wrote %d bytes to '%s'.", bytesWritten, destinationPath)

	// Verify Integrity (SHA256 Hash of written file)
	// We hash rawFileBytes because that's what the server hashed and sent (after base64).
	// If we read back from disk, it's a double check on the OS write, but for direct upload,
	// hashing the bytes we intended to write is a valid verification of the data *received*.
	hasher := sha256.New()
	hasher.Write(rawFileBytes)
	actualSha256 := hex.EncodeToString(hasher.Sum(nil))

	hashMatched := actualSha256 == args.ExpectedSha256
	finalMessage := fmt.Sprintf("File '%s' uploaded to '%s' (%d bytes). Expected SHA256: %s, Actual SHA256: %s. Hash Verified: %t.",
		cleanFilename, args.TargetDirectory, bytesWritten, args.ExpectedSha256, actualSha256, hashMatched)

	if !hashMatched {
		log.Printf("|❗UPLOAD WARNING HASH MISMATCH for %s.", destinationPath)
		return models.UploadResult{
			FilePath:     destinationPath,
			ActualSha256: actualSha256,
			Message:      finalMessage,
			HashMatched:  false,
		}, fmt.Errorf("hash verification failed")
	}

	log.Printf("|👊 UPLOAD SUCCESS| Upload successful and hash VERIFIED for %s.", destinationPath)
	return models.UploadResult{
		FilePath:     destinationPath,
		ActualSha256: actualSha256,
		Message:      finalMessage,
		HashMatched:  true,
	}, nil

}
