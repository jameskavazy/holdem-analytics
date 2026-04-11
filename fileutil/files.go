package fileutil

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

// TODO: Build a separate package to handle filesystem interactions e.g. fsnotify watcher and moving processed files around...

// MoveProcessedFiles moves a file to specified destination under newPath. It first checks that the directory in newPath exists, and creates it if required.
func MoveProcessedFiles(oldPath, newPath string) error {
	found, err := checkDirExists(filepath.Dir(newPath))
	if err != nil {
		log.Printf("an unexpected error occurred with the specied path")
		return err
	}

	if !found {
		mkdirErr := os.Mkdir(filepath.Dir(newPath), 0750)
		if mkdirErr != nil {
			log.Println("could not create folder for processed hands")
			return mkdirErr
		}
	}

	return moveFile(oldPath, newPath)
}

func moveFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func checkDirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return false, errors.New("specified path " + path + " is not a directory")
		}
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
