package fileutil

import "os"

// TODO: Build a separate package to handle filesystem interactions e.g. fsnotify watcher and moving processed files around...

func MoveFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
