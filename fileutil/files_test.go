package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveProcessedFiles(t *testing.T) {

	oldPath := "testing-123.txt"
	newPath := "test_processed/testing-123.txt"

	file, createErr := os.Create(oldPath)
	if createErr != nil {
		t.Fatalf("test setup failed: %v", createErr)
	}
	closeErr := file.Close()
	if closeErr != nil {
		t.Fatalf("test set up failed %v", closeErr)
	}

	t.Cleanup(func() {
		os.RemoveAll(filepath.Dir(newPath))
	})

	err := MoveProcessedFiles(oldPath, newPath)

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s but it wasn't there", newPath)
	}

	if err != nil {
		t.Errorf("error %#v detected but didn't expect one", err.Error())
	}

}

func TestCheckDirExists(t *testing.T) {

	t.Run("the dir doesn't exist yet", func(t *testing.T) {

		newPath := "test_processed/testing-123.txt"

		found, err := checkDirExists(filepath.Dir(newPath))

		if found {
			t.Errorf("expected %s not to be found but it was", newPath)
		}

		if err != nil {
			t.Errorf("wanted no error but got %v", err)
		}

	})

	t.Run("the dir already exists", func(t *testing.T) {
		os.Mkdir("check_dir_exists_already_exists", 0750)
		newPath := "check_dir_exists_already_exists/testing-123.txt"
		t.Cleanup(func() {
			os.RemoveAll(filepath.Dir(newPath))
		})

		found, err := checkDirExists(filepath.Dir(newPath))

		if !found {
			t.Errorf("wanted found=true but got found=false and err: %v", err)
		}

		if err != nil {
			t.Errorf("wanted nil error but got %v", err)
		}

	})
}
