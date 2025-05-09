package notify

import (
	"io/fs"
	"path/filepath"

	"github.com/fatih/color"
)

// HandleNewDirectory handles the new directory event. Recursively adds new directories to the watcher if not ignored.
// A directory check should be performed prior to calling this function to ensure the path is a directory, not a file.
func HandleNewDirectory(path string, w *Watcher) {
	if path == "" {
		return
	}

	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		// skip on error, non directory path, or ignored directory path
		if err != nil || !d.IsDir() || w.ShouldIgnoreDir(path) {
			return nil
		}

		// add the path to the watcher
		if err := w.Add(path); err != nil {
			color.Red("failed to watch %s with error %v", path, err)
		}

		return nil
	})
}
