package pokerhud_test

import (
	"os"
	"pokerhud/hands"
	"testing"
)

func BenchmarkHandHistoryFromFS(b *testing.B) {
	fileSystem := os.DirFS("C:\\Users\\james\\testfolder")
	for i := 0; i < b.N; i++ {
		hands.ExportHands(fileSystem)
		// hands.ExportHands(fstest.MapFS{
		// 	"zoom.txt":      {Data: []byte(zoomHand1)},
		// 	"cash game.txt": {Data: []byte(cashGame1)},
		// 	"failure.txt":   {Data: []byte("not a hand")},
		// })
	}
}
