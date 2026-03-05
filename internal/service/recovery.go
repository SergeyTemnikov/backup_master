package service

import (
	"os"
	"path/filepath"
	"strings"
)

func CleanupTempArtifacts(root string) error {

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".tmp") {

			if info.IsDir() {
				return os.RemoveAll(path)
			}

			return os.Remove(path)
		}

		return nil
	})
}
