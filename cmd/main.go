package main

import (
	"fmt"
	"os"
	"pokerhud"
)

func main() {
	// fileSystem := os.DirFS("C:\\Users\\james\\testfolder")
	fileSystem := os.DirFS("C:\\Users\\james\\AppData\\Local\\PokerStars.UK\\HandHistory\\KavarzE")
	hands, handErrs := pokerhud.HandHistoryFromFS(fileSystem)

	fmt.Printf("%#v\n", hands[0])
	fmt.Println("Hands: ", len(hands))
	fmt.Println("Errors: ", len(handErrs))
	
}
