//go:build linux
// +build linux

package platform

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Startup returns the system startup time on Linux systems by using the who -b command
// Startup returns the system boot time constructed from the output of the external
// command "who -b".
//
// Behavior:
//   - Executes the command "who -b" and reads its stdout. If the command fails the
//     returned error is non-nil and the zero time is returned.
//   - The implementation expects the command output to contain a time component at a
//     fixed offset and slices the output (skipping the date portion) to extract an
//     "HH:MM" string, then parses hours and minutes with strconv.Atoi.
//   - The returned time.Time is built using the current year, month and day (time.Now()),
//     the parsed hour and minute, zero seconds and nanoseconds, and the current local
//     location (now.Location()).
//
// Important caveats and limitations:
//   - This function is platform- and output-format dependent (relies on "who -b" and a
//     specific output layout) and is not robust to variations in that output.
//   - The code ignores parsing errors for hours/minutes (strconv.Atoi errors are discarded);
//     if parsing fails the hour and/or minute default to zero and the function will return
//     a time on the current date at 00:00 with a nil error.
//   - The date portion of the boot time is intentionally skipped: the function uses today's
//     date rather than the actual boot date, which can produce incorrect results for boots
//     that occurred on a previous day (e.g., across midnight) or when the system clock has
//     changed.
//   - This approach may not work in restricted environments (missing "who" binary, PATH
//     differences, containers) and should be used with caution. Consider using a more
//     robust method (e.g., parsing /proc/uptime or using system APIs) for production code.
func Startup() (time.Time, error) {
	cmd := exec.Command("who", "-b")
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}
	// Dodgy and dangerous - we skip the date part
	startupTimeStr := strings.TrimSpace(string(output))[24:]
	hours, _ := strconv.Atoi(startupTimeStr[0:2])
	minutes, _ := strconv.Atoi(startupTimeStr[3:5])
	now := time.Now()

	// Clean up the output by removing newlines and extra spaces
	return time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, now.Location()), nil
}
