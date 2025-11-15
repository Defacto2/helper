//go:build windows

package helper

import (
	"errors"
	"fmt"
)

var ErrWindows = errors.New("this func is not supported on windows")

// DiskStat returns the total, free, and percentage free of the drive path.
// If no path is provided then the drive of the working directory is used.
//
// It is current unsupported on Windows and always returns an error.
func DiskStat(path string) (float64, float64, float64, string, error) {
	return 0, 0, 0, "", fmt.Errorf("disk stat getwd: %w", ErrWindows)
}
