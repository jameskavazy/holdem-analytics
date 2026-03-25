package main

import (
	"fmt"
	"os"
	"pokerhud"
	"time"
)

func main() {
	start := time.Now()

	fileSystem := os.DirFS("C:\\Users\\james\\testfolder")
	// fileSystem := os.DirFS("C:\\Users\\james\\AppData\\Local\\PokerStars.UK\\HandHistory\\KavarzE")

	hands, handErrs := pokerhud.HandHistoryFromFS(fileSystem)

	elapsed := time.Since(start)
	fmt.Printf("Processed %d hands in %s\n", len(hands), elapsed)

	fmt.Printf("%#v\n", hands[0])
	fmt.Println("Hands: ", len(hands))
	fmt.Println("Errors: ", len(handErrs))
}
