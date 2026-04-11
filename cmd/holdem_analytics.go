package main

import (
	"log"
	"os"
	"path/filepath"
	"pokerhud/fileutil"
	"pokerhud/hands"
)

func main() {
	args := os.Args
	argsLen := len(args)
	if argsLen > 2 {
		log.Fatal("too many arguments provided.\n example usage: ./holdem-analytics <hand history folder path>")
	} else if argsLen < 2 {
		log.Fatal("not enough arguments provided.\n example usage: ./holdem-analytics <hand history folder path>")
	}

	targetDir := args[1]
	proccessedDir := filepath.Join(targetDir, "Processed By Holdem Analytics")

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
