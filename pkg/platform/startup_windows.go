//go:build windows
// +build windows

package platform

import (
	"os/exec"
	"strings"
)

func Startup() (string, error) {
	cmd := exec.Command("powershell", "-Command", " (Get-EventLog -LogName System -After (Get-Date -Hour 0 -Minute 0 -Second 0 -Millisecond 0) | Select-Object -Last 1).TimeGenerated.ToString(\"HH:mm\")")
	output, err := cmd.CombinedOutput()
	outputStr := ""
	if err == nil {
		outputStr = strings.Trim(string(output), "\r\n")
	}
	return outputStr, err
}
