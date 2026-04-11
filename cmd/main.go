package main

import (
	"log"
	"os"
	"path/filepath"
	"pokerhud/fileutil"
	"pokerhud/hands"
)

func main() {
	targetDir := filepath.Join("C:", "Users", "james", "testfolder")
	proccessedDir := filepath.Join(targetDir, "processed")

	fileSystem := os.DirFS(targetDir)

	result := hands.ExportHands(fileSystem)

	for _, f := range result.SuccessFiles() {
		oldPath := filepath.Join(targetDir, f)
		newPath := filepath.Join(proccessedDir, f)

		err := fileutil.MoveProcessedFiles(oldPath, newPath)

		if err != nil {
			log.Printf("error moving file %s", err.Error())
		}
	}

	log.Printf(
		"Successful Files: %#v\nFailedFiles: %#v\nFsError: %#v", result.SuccessCount(), result.FileErrorCount(), result.FsErr,
	)

	log.Printf("Hands parsed: %v", result.HandsCount())
	log.Printf("Hand errs: %v", result.HandErrCount())

}
