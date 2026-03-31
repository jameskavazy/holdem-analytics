package main

import (
	"log"
	"os"
	"pokerhud/hands"
)

func main() {
	fileSystem := os.DirFS("C:\\Users\\james\\testfolder")
	// fileSystem := os.DirFS("C:\\Users\\james\\AppData\\Local\\PokerStars.UK\\HandHistory\\KavarzE")
	// fileSystem := os.DirFS("C:\\Coding\\pokerhud")

	result := hands.ExportHands(fileSystem)

	log.Printf(
		"Successful Files: %#v\nFailedFiles: %#v\nFsError: %#v", result.SuccessCount(), result.FileErrorCount(), result.FsErr,
	)

	log.Printf("Hands parsed: %v", result.HandsCount())
	log.Printf("Hand errs: %v", result.HandErrCount())

}
