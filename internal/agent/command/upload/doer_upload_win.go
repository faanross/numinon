//go:build windows

package upload

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// windowsUpload implements the CommandUpload interface for Windows.
type windowsUpload struct{}

// New is the constructor for our Windows-specific Upload command
func New() CommandUpload {
	return &windowsUpload{}
}

// DoUpload executes the file upload logic.
func (wu *windowsUpload) DoUpload(args models.UploadArgs) (models.UploadResult, error) {

	// Display command arguments to terminal (TODO convert to slog)
	fmt.Println("|✅ UPLOAD DOER| The UPLOAD command has been executed.")
	fmt.Printf("|📋 UPLOAD DETAILS| TargetDir: '%s', TargetFilename: '%s', Overwrite: %t, ExpectedSHA256: %s, ContentLen(b64): %d\n",
		args.TargetDirectory, args.TargetFilename, args.OverwriteIfExists, args.ExpectedSha256, len(args.FileContentBase64))

	// Validate command-specific input
	if args.TargetDirectory == "" || args.TargetFilename == "" || args.FileContentBase64 == "" || args.ExpectedSha256 == "" {
		return nil, fmt.Errorf("missing essential arguments for upload (TargetDirectory, TargetFilename, FileContentBase64, ExpectedSha256)")
	}

	// Decode File Content
	rawFileBytes, err := base64.StdEncoding.DecodeString(args.FileContentBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode file content: %w", err)
	}
	log.Printf("|⚙️ UPLOAD ACTION| Decoded file content: %d bytes\n", len(rawFileBytes))

	// Construct Full Destination Path
	// Ensure TargetDirectory exists
	// TODO if not create it, for now, assume it must exist or OS handles error
	// TODO for OPSEC sanitation, ensure TargetDirectory is within an allowed base path
	destinationPath := filepath.Join(args.TargetDirectory, args.TargetFilename)
	log.Printf("|⚙️ UPLOAD ACTION| Full destination path: %s\n", destinationPath)

	// Check for Existing File & Overwrite Logic
	_, statErr := os.Stat(destinationPath)
	if statErr == nil { // File exists
		if !args.OverwriteIfExists {
			return nil, fmt.Errorf("file '%s' already exists and overwrite is not permitted", destinationPath)
		}
		log.Printf("|⚙️ UPLOAD ACTION| File '%s' exists and will be overwritten\n", destinationPath)
	} else if !os.IsNotExist(statErr) { // Some other error trying to stat the file
		return nil, fmt.Errorf("error checking destination file '%s': %w", destinationPath, statErr)
	}

	// Write File to Disk
	// Using 0600 for more restrictive permissions initially.
	err = os.WriteFile(destinationPath, rawFileBytes, 0600)
	if err != nil {
		// TODO: Differentiate common errors like permission denied, disk full
		return nil, fmt.Errorf("failed to write file to '%s': %w", destinationPath, err)
	}
	log.Printf("|⚙️ UPLOAD ACTION| Successfully wrote %d bytes to '%s'", len(rawFileBytes), destinationPath)

	// Verify Integrity (SHA256 Hash of written file)
	// For maximum integrity, read back the file just written to ensure what's on disk is what we hash.
	// If performance is a concern for huge files AND we trust os.WriteFile,
	// we could hash rawFileBytes directly, but reading back is safer.
	writtenBytes, err := os.ReadFile(destinationPath)
	if err != nil {
		// This is problematic: file written but can't be read back for verification.
		return &UploadResult{ // Still return some info
			FilePath:     destinationPath,
			BytesWritten: int64(len(rawFileBytes)), // Assumed written length
			Message:      fmt.Sprintf("File written to '%s', but failed to read back for verification: %v. HASH UNVERIFIED.", destinationPath, err),
		}, fmt.Errorf("file written but could not be re-read for hash verification: %w", err)
	}

	hasher := sha256.New()
	// It's crucial that what we hash here is exactly what the server hashed.
	// The server sent hash of original (decoded) file content.
	// If os.WriteFile is atomic and successful, rawFileBytes should equal writtenBytes.
	// Let's hash rawFileBytes as that's what server sent (after decoding).
	// To be absolutely sure about what's on disk vs expected:
	hasher.Write(writtenBytes) // Hash what was actually read from disk
	actualSha256 := hex.EncodeToString(hasher.Sum(nil))

	resultMsg := fmt.Sprintf("File '%s' uploaded to '%s' (%d bytes). Expected SHA256: %s, Actual SHA256: %s.",
		args.TargetFilename, args.TargetDirectory, len(writtenBytes), args.ExpectedSha256, actualSha256)

	uploadRes := &UploadResult{
		FilePath:     destinationPath,
		BytesWritten: int64(len(writtenBytes)),
		ActualSha256: actualSha256,
		Message:      resultMsg,
	}

	if actualSha256 != args.ExpectedSha256 {
		log.Printf("|❗UPLOAD WARNING| HASH MISMATCH for %s. Expected: %s, Got: %s", destinationPath, args.ExpectedSha256, actualSha256)
		return uploadRes, fmt.Errorf("hash mismatch after write") // Signal error for mismatch
	}

	log.Printf("|👊 UPLOAD SUCCESS| Upload successful and hash VERIFIED for %s.", destinationPath)
	return uploadRes, nil

}
