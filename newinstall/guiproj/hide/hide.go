// +build !windows

package hide

import (
	"os"
	"path/filepath"
	"strings"
)

func HideFile(filename string) error {
	if !strings.HasPrefix(filepath.Base(filename), ".") {
		dir := filepath.Dir(filename)
		base := filepath.Base(filename)
		base = "." + base
		err := os.Rename(filename, filepath.Join(dir, base))
		if err != nil {
			return err
		}
	}
	return nil
}
