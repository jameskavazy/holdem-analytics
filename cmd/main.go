package main

import (
	"fmt"
	"os"
	"pokerhud"
)

func main() {
	// fileSystem := os.DirFS("C:\\Users\\james\\testfolder")
	fileSystem := os.DirFS("C:\\Users\\james\\AppData\\Local\\PokerStars.UK\\HandHistory\\KavarzE")
	hands, err := pokerhud.HandHistoryFromFS(fileSystem)
	fmt.Println(hands[0])
	fmt.Println(len(hands))
	fmt.Println("Errors: ", len(err))
	for _, e := range err {
		fmt.Printf(e.Error(), "\n\n")
	}
}
