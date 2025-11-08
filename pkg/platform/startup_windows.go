//go:build windows
// +build windows

package platform

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Startup retrieves the system startup time on Windows by querying the System EventLog.
// It executes a PowerShell command to get the last event log entry's timestamp from the current day.
// The function returns a time.Time object representing the startup time and an error.
//
// The returned time will have the current date but with hours and minutes from the startup event.
// Seconds and nanoseconds are set to 0.
//
// Note: This implementation has limitations as it:
// - Only works on Windows systems
// - Requires PowerShell to be available
// - Assumes the last event log entry corresponds to startup
// - Ignores potential errors from time parsing
//
// Returns:
//   - time.Time: The system startup time with current date
//   - error: Any error encountered during execution of the PowerShell command
func Startup() (time.Time, error) {
	cmd := exec.Command("powershell", "-Command", " (Get-EventLog -LogName System -After (Get-Date -Hour 0 -Minute 0 -Second 0 -Millisecond 0) | Select-Object -Last 1).TimeGenerated.ToString(\"HH:mm\")")
	output, err := cmd.CombinedOutput()
	outputStr := ""
	if err == nil {
		outputStr = strings.Trim(string(output), "\r\n")
	}

	// Dodgy and dangerous - we skip the date part
	hours, _ := strconv.Atoi(outputStr[0:2])
	minutes, _ := strconv.Atoi(outputStr[3:5])
	now := time.Now()

	// Clean up the output by removing newlines and extra spaces
	return time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, now.Location()), nil
}
