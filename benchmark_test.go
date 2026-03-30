package pokerhud_test

// import (
// 	// "os"
// 	"pokerhud"
// 	"testing"
// 	"testing/fstest"
// )

// func BenchmarkHandHistoryFromFS(b *testing.B) {
// 	// fileSystem := os.DirFS("C:\\Users\\james\\AppData\\Local\\PokerStars.UK\\HandHistory\\KavarzE")
// 	for i := 0; i < b.N; i++ {
// 		// pokerhud.HandHistoryFromFS(fileSystem)
// 		pokerhud.ExportHands(fstest.MapFS{
// 			"zoom.txt":      {Data: []byte(zoomHand1)},
// 			"cash game.txt": {Data: []byte(cashGame1)},
// 			"failure.txt":   {Data: []byte("not a hand")},
// 		})
// 	}

// 	// pokerhud.HandHistoryFromFS(fstest.MapFS{
// 	// 	"zoom.txt":      {Data: []byte(zoomHand1)},
// 	// 	"cash game.txt": {Data: []byte(cashGame1)},
// 	// 	"failure.txt":   {Data: []byte("not a hand")},
// 	// })
// }
