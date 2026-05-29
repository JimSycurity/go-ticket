//go:build !windows

package ticket

import "os"

func replaceFile(src string, dst string) error {
	return os.Rename(src, dst)
}
