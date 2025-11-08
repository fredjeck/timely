//go:build !windows && !linux
// +build !windows,!linux

package platform

import (
	"fmt"
	"time"
)

func Startup() (time.Time, error) {
	return time.Time{}, fmt.Errorf("Startup function not implemented for this platform")
}
