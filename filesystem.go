package bingo

import (
	"os"
)

// setupFolder creates a folder (including subfolders) if it does not exist already.
func setupFolder(folder string, perm os.FileMode) error {
	// Create folder if it does not exist yet
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		Loggers.Info.Printf("Create folder %s", folder)
		if mkdirerr := os.MkdirAll(folder, perm); mkdirerr != nil {
			return mkdirerr
		}
	}
	return nil
}
