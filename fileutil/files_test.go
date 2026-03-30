package fileutil

import (
	// "fmt"
	"os"
	"testing"
)

func TestMoveFile(t *testing.T) {
	
	oldPath := "testing-123.txt"
	newPath := "test_processed/testing-123.txt"

	file, _ := os.Create(oldPath)
	file.Close()
	
	err := MoveFile(oldPath, newPath)

	if err != nil {
		t.Errorf("error %#v detected but didn't expect one", err.Error())
	}
}