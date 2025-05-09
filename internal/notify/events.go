package notify

import (
	"errors"
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

	// recursively add to watcher if not ignored.
	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		// skip on error, non directory path
		if err != nil || !d.IsDir() {
			return nil
		}

		// skip if ignored directory
		if w.ShouldIgnoreDir(path) {
			return fs.SkipDir
		}

		// add the path to the watcher
		if err := w.Add(path); err != nil {
			color.Red("failed to watch %s with error %v", path, err)
		} else {
			color.Magenta("watching %s", path)
		}

		return nil
	})
}

// HandleRemoveDirectory handles the remove directory event. Recursively removes any watched subdirectories.
// A directory check should be performed prior to calling this function to ensure the path is a directory, not a file.
func HandleRemoveDirectory(path string, w *Watcher) {
	if path == "" {
		return
	}

	// subdirectories are automatically removed if watched
	if err := w.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		color.Red("failed to unwatch %s with error %v", path, err)
	} else {
		color.Magenta("unwatched %s", path)
	}
}
