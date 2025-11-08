//go:build !windows
// +build !windows

package platform

import "fmt"

func Startup() (string, error) {
	return "", fmt.Errorf("Startup function not implemented for this platform")
}
