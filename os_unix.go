//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

//nolint:nonamedreturns
package helper

import (
	"fmt"
	"math"
	"os"

	"golang.org/x/sys/unix"
)

// DiskStat returns the total bytes, free bytes, percentage free, and formatted percentage string.
// If no path is provided, the drive of the current working directory is used.
func DiskStat(path string) (
	total float64, free float64, percentage float64, formatted string, err error,
) {
	var stat unix.Statfs_t
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return 0, 0, 0, "", fmt.Errorf("disk stat getwd: %w", err)
		}
		path = wd
	}
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, 0, 0, "", fmt.Errorf("unix stat fs: %w", err)
	}

	// Use fragment size (Frsize) for accurate block allocation math
	bsize := float64(stat.Frsize)
	if bsize == 0 {
		bsize = float64(stat.Bsize) // fallback if Frsize is not reported
	}

	total = float64(stat.Blocks) * bsize
	free = float64(stat.Bavail) * bsize

	if total == 0 {
		return 0, 0, 0, "0%", nil
	}

	const calc = 100.0
	x := (free / total) * calc
	s := fmt.Sprintf("%d%%", int64(math.Round(x)))

	return total, free, x, s, nil
}
