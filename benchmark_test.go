package pokerhud_test

import (
	"os"
	"pokerhud/hands"
	"testing"
)

func BenchmarkHandHistoryFromFS(b *testing.B) {
	fileSystem := os.DirFS("C:\\Users\\james\\testfolder")

	for b.Loop() {
		hands.ExportHands(fileSystem)
	}
}
