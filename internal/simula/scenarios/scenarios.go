package scenarios

import "math/rand"

// --- Run Command Scenarios ---
// Collections of shell commands categorized by adversary intent.

var SystemReconCommands = []string{
	"hostname",
	"systeminfo",
	"systeminfo | findstr /B /C:\"OS Name\" /C:\"OS Version\" /C:\"System Type\"",
	"echo %PROCESSOR_ARCHITECTURE%",
	"wmic cpu get name, maxclockspeed, numberofcores",
	"wmic computersystem get model,manufacturer",
	"reg query HKLM\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall",
	"schtasks /query /fo LIST /v",
}

var NetworkReconCommands = []string{
	"ipconfig /all",
	"netstat -anob",
	"arp -a",
	"nbtstat -n",
	"netsh wlan show profiles",
	"netsh interface show interface",
	"route print",
	"nltest /dclist:", // Note: Will fail on non-domain joined machines
	"nslookup -type=SRV _ldap._tcp.dc._msdcs.", // Note: Will fail on non-domain joined
}

var UserReconCommands = []string{
	"whoami",
	"whoami /all",
	"whoami /groups",
	"echo %USERNAME%",
	"quser",
	"query user",
	"qwinsta",
	"for /F \"tokens=*\" %U in ('dir /b /ad \"C:\\Users\"') do (dir /b \"C:\\Users\\%U\\Desktop\")",
}

var DiscoveryCommands = []string{
	"net user",
	"net accounts",
	"net localgroup",
	"net localgroup administrators",
	"net group \"Domain Admins\" /domain",     // Note: Will fail on non-domain joined
	"net group \"Enterprise Admins\" /domain", // Note: Will fail on non-domain joined
	"tasklist /v",
	"dir C:\\Users\\",
	"dir C:\\",
	"dir C:\\ProgramData",
	"dir \"C:\\Program Files\"",
	"sc query",
}

// --- Upload Scenarios ---
// Defines a preset for uploading a tool. It links a dummy file to
// plausible destination directories on the target host.
type UploadPreset struct {
	DummyFileName       string
	PlausibleTargetDirs []string
}

var ToolingUploads = []UploadPreset{
	{DummyFileName: "procdump64.exe", PlausibleTargetDirs: []string{"C:\\Windows\\Temp", "C:\\PerfLogs"}},
	{DummyFileName: "psexec.exe", PlausibleTargetDirs: []string{"C:\\Windows\\System32", "C:\\Temp"}},
	{DummyFileName: "svc-utility.exe", PlausibleTargetDirs: []string{"C:\\Users\\Public\\", "C:\\Temp"}},
	{DummyFileName: "autoruns.exe", PlausibleTargetDirs: []string{"C:\\Temp", "C:\\Users\\Public\\Downloads"}},
	{DummyFileName: "pwsh.exe", PlausibleTargetDirs: []string{"C:\\ProgramData\\PowerShell", "C:\\Windows\\Temp"}},
	{DummyFileName: "wevtutil.exe", PlausibleTargetDirs: []string{"C:\\Temp\\SysUtils"}},
	{DummyFileName: "enum-local.bat", PlausibleTargetDirs: []string{"C:\\Users\\Public\\Documents", "C:\\Temp"}},
	{DummyFileName: "Find-NetworkShares.ps1", PlausibleTargetDirs: []string{"C:\\Windows\\Temp", "C:\\Users\\Public\\Scripts"}},
	{DummyFileName: "Get-ADUsers.ps1", PlausibleTargetDirs: []string{"C:\\Windows\\Temp", "C:\\Users\\Public\\Scripts"}},
	{DummyFileName: "logon-helper.exe", PlausibleTargetDirs: []string{"C:\\ProgramData\\Microsoft", "C:\\Temp"}},
	{DummyFileName: "kb-hook.dll", PlausibleTargetDirs: []string{"C:\\Windows\\System32", "C:\\Program Files\\Common Files\\System"}},
	{DummyFileName: "settings.xml", PlausibleTargetDirs: []string{"C:\\ProgramData", "C:\\Users\\Public\\"}},
	{DummyFileName: "AdobeUpdate.exe", PlausibleTargetDirs: []string{"C:\\Users\\Public\\Downloads", "C:\\ProgramData\\Adobe"}},
	{DummyFileName: "jucheck.exe", PlausibleTargetDirs: []string{"C:\\ProgramData\\Oracle\\Java", "C:\\Windows\\Temp"}},
	{DummyFileName: "backupsvc.exe", PlausibleTargetDirs: []string{"C:\\Program Files\\Common Files", "C:\\Windows\\SysWOW64"}},
	{DummyFileName: "ms-netlib.dll", PlausibleTargetDirs: []string{"C:\\Windows\\System32"}},
	{DummyFileName: "config.json", PlausibleTargetDirs: []string{"C:\\ProgramData", "C:\\Temp"}},
	{DummyFileName: "archive.zip", PlausibleTargetDirs: []string{"C:\\Users\\Public\\", "C:\\PerfLogs"}},
	{DummyFileName: "network.txt", PlausibleTargetDirs: []string{"C:\\Users\\Public\\Documents", "C:\\Temp"}},
}

// --- Download Scenarios ---
// Lists of plausible file paths for data exfiltration.
var ExfilSystemFiles = []string{
	"C:\\Windows\\System32\\drivers\\etc\\hosts",
	"C:\\Windows\\System32\\config\\SAM",    // Note: Will fail without high privileges
	"C:\\Windows\\System32\\config\\SYSTEM", // Note: Will fail without high privileges
	"C:\\Windows\\security\\logs\\scesetup.log",
	"C:\\Windows\\PFRO.log",
	"C:\\Windows\\debug\\NetSetup.log",
}
var ExfilUserFiles = []string{
	// NOTE: Environment variables are not yet expanded by the agent.
	// These paths are treated as literals for now.
	"C:\\Users\\Public\\Documents\\Financials.xlsx",
	"C:\\Users\\Public\\Desktop\\credentials.txt",
	"C:\\Users\\Public\\Desktop\\network-diagram.vsdx",
	"C:\\Users\\Public\\Downloads\\archive.zip",
}

// --- Enumerate Scenarios ---
// Lists of process names to check for, categorized by type.
var SecurityProductProcesses = []string{"MsMpEng.exe", "SentinelAgent.exe", "carbonblack.exe", "cb.exe", "NisSrv.exe", "SAVService.exe", "TmListen.exe", "WinDefend.exe"}
var RemoteAccessProcesses = []string{"mstsc.exe", "TeamViewer.exe", "AnyDesk.exe", "chrome.exe", "svchost.exe -k termsvcs", "winvnc.exe"}

// GetRandom selects a random string from a slice of strings.
func GetRandom(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[rand.Intn(len(slice))]
}
