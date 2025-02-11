package main

import (
	// "fmt"
	// "os"
	// "pokerhud"
	"fmt"
	"slices"
)

func main() {
	// fileSystem := os.DirFS("C:\\Users\\james\\testfolder")
	// fileSystem := os.DirFS("C:\\Users\\james\\AppData\\Local\\PokerStars.UK\\HandHistory\\KavarzE")
	// hands, _ := pokerhud.HandHistoryFromFS(fileSystem)
	// fmt.Println(hands[0])
	// fmt.Println(hands[55])

	arr := []string{"hi", "woops", "ho"}
	i := slices.IndexFunc(arr, func(s string) bool {
		return s == "nljn"
	})

	fmt.Println(i)
}
