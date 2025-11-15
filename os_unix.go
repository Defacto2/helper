//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

package helper

import (
	"fmt"
	"math"
	"os"

	"golang.org/x/sys/unix"
)

// DiskStat returns the total, free, and percentage free of the drive path.
// If no path is provided then the drive of the working directory is used.
//
// The returned string is a humanized percentage, ie "50%".
func DiskStat(path string) (float64, float64, float64, string, error) {
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
	// Available blocks * size per block = available space in bytes
	totl := float64(stat.Blocks) * float64(stat.Bsize)
	free := float64(stat.Bavail) * float64(stat.Bsize)
	perc := (free / totl) * 100
	s := fmt.Sprintf("%d%%", int64(math.Floor(perc+0.5)))
	return totl, free, perc, s, nil
}
