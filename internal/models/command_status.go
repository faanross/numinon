package models

// Success Constants - GENERAL
const (
	StatusSuccess             = "success"
	StatusSuccessHashVerified = "success_hash_verified"
	StatusSuccessHashMismatch = "success_hash_mismatch" // File written, but integrity issue
	StatusSuccessLaunched     = "success_launched"      // For shellcode
	StatusSuccessHopInitiated = "success_hop_initiated"
	StatusSuccessMorphApplied = "success_morph_applied"
	StatusSuccessMorphPartial = "success_morph_partial"
	StatusSuccessExitNonZero  = "success_exit_non_zero" // Command ran but returned non-zero

)

// Failure Constants

const (
	StatusFailureInvalidArgs           = "failure_invalid_args"
	StatusFailureDecodeError           = "failure_decode_error"
	StatusFailureFileExistsNoOverwrite = "failure_file_exists_no_overwrite"
	StatusFailureWriteError            = "failure_write_error"
	StatusFailureReadError             = "failure_read_error"
	StatusFailurePermissionDenied      = "failure_permission_denied"
	StatusFailureVerificationError     = "failure_verification_error" // e.g. couldn't re-read for hash
	StatusFailureExecutionError        = "failure_execution_error"    // General execution failure
	StatusFailureFileNotFound          = "failure_file_not_found"     // For download function could not find source file
	StatusFailureLoaderError           = "failure_loader_error"
	StatusFailureMorphNoValidChanges   = "failure_morph_no_valid_changes"
	StatusFailureTimeout               = "failure_timeout"
	StatusFailureExecError             = "failure_exec_error"
	StatusFailureUnmarshallError       = "Failed to unmarshall JSON object"
	StatusFailureUnknownCommand        = "The command is unknown - does not exists, or is not currently registered on agent"
	StatusFailureNotSupported          = "This feature is not yet supported on this agent"
)
