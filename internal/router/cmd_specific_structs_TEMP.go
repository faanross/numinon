package router

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"numinon_shadow/internal/models"
	"os"
)

func returnUploadStruct(w http.ResponseWriter) []byte {

	// HERE WE CREATE response.Data, UPLOAD SPECIFIC ARGUMENTS
	fileBytes, err := os.ReadFile("./dummy/dummy.txt")
	if err != nil {
		panic(fmt.Errorf("failed to read prerequisite file: %w", err))
	}
	hashBytes := sha256.Sum256(fileBytes)

	uploadArguments := models.UploadArgs{
		TargetDirectory:   "C:\\Users\\vuilhond\\Desktop\\",
		TargetFilename:    "dummy.txt",
		FileContentBase64: base64.StdEncoding.EncodeToString(fileBytes),
		ExpectedSha256:    fmt.Sprintf("%x", hashBytes),
		OverwriteIfExists: true,
	}

	uploadArgsJSON, err := json.Marshal(uploadArguments)
	if err != nil {
		log.Printf("Failed to marshal upload args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return uploadArgsJSON
}

func returnDownloadStruct(w http.ResponseWriter) []byte {

	downloadArguments := models.DownloadArgs{
		SourceFilePath: "C:\\Users\\vuilhond\\Desktop\\download_me.txt",
	}

	downloadArgsJSON, err := json.Marshal(downloadArguments)
	if err != nil {
		log.Printf("Failed to marshal download args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return downloadArgsJSON

}

func returnRunCmdStruct(w http.ResponseWriter) []byte {

	runCmdArguments := models.RunCmdArgs{
		CommandLine: "whoami",
	}

	runCmdArgsJSON, err := json.Marshal(runCmdArguments)
	if err != nil {
		log.Printf("Failed to marshal runcmd args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return runCmdArgsJSON

}
