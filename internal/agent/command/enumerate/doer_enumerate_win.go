//go:build windows

package enumerate

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/faanross/numinon/internal/models"
	"golang.org/x/sys/windows"
)

// windowsEnumerate implements the CommandEnumerate interface for Windows.
type windowsEnumerate struct{}

// New is the constructor for our Windows-specific Enumerate command
func New() CommandEnumerate {
	return &windowsEnumerate{}
}

// DoEnumerate performs the actual endpoint enumeration on Windows-systems
func (we *windowsEnumerate) DoEnumerate(args models.EnumerateArgs) (models.EnumerateResult, error) {
	fmt.Println("|✅ ENUMERATE DOER| The ENUMERATE command has been executed.")
	var processes []models.ProcessInfo

	snapshotHandle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		msg := fmt.Sprintf("Windows: CreateToolhelp32Snapshot failed: %v", err)
		log.Printf("|❗ERR ENUMERATE DOER| %s", msg)
		return models.EnumerateResult{Message: msg}, fmt.Errorf(msg)
	}
	defer windows.CloseHandle(snapshotHandle)

	var entry windows.ProcessEntry32 // Using ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshotHandle, &entry) // Using Process32First
	if err != nil {
		msg := fmt.Sprintf("Windows: Process32First failed: %v", err)
		log.Printf("|❗ERR ENUMERATE DOER| %s", msg)
		return models.EnumerateResult{Message: msg}, fmt.Errorf(msg)
	}

	foundCount := 0
	for {
		// ExeFile is [MAX_PATH]uint16, so UTF16ToString is correct
		processName := windows.UTF16ToString(entry.ExeFile[:])
		pid := entry.ProcessID

		if args.ProcessName == "" || strings.EqualFold(processName, args.ProcessName) {
			processes = append(processes, models.ProcessInfo{
				PID:  pid,
				Name: processName,
			})
			foundCount++
		}

		err = windows.Process32Next(snapshotHandle, &entry) // Using Process32Next
		if err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			msg := fmt.Sprintf("Windows: Process32Next failed: %v", err)
			log.Printf("|❗ERR ENUMERATE DOER| %s", msg)
			return models.EnumerateResult{Processes: processes, Message: msg}, fmt.Errorf(msg)
		}
	}

	successMsg := fmt.Sprintf("Windows: Successfully enumerated %d process(es). Filter: '%s'", len(processes), args.ProcessName)
	log.Printf("|✅ ENUMERATE DOER | %s", successMsg)
	return models.EnumerateResult{Processes: processes, Message: successMsg}, nil

}
