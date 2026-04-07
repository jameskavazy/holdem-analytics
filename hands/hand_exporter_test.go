package hands

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestStreamHands(t *testing.T) {
	fileSystem := fstest.MapFS{
		"zoom.txt": {Data: []byte(testHands)},
	}

	dir, _ := fs.ReadDir(fileSystem, ".")

	hands := streamHands(fileSystem, dir)

	count := 0
	for range hands {
		count++
	}

	if count != 1 {
		t.Errorf("wanted 1 hand but got %d", count)
	}
}

func TestCollectResults(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"zoom.txt": {Data: []byte(testHands)},
		}
		dir, _ := fs.ReadDir(fileSystem, ".")

		handsChannel := streamHands(fileSystem, dir)

		got := collectResults(handsChannel)

		successCount, failureCount := sumHandsHelper(got.FileResults)

		if successCount != 1 {
			t.Errorf("wanted 1 successCount but got %d", successCount)
		}

		if failureCount != 0 {
			t.Errorf("wanted 0 successCount but got %d", failureCount)
		}

		s := got.FileResults[0]
		if s.Path != "zoom.txt" {
			t.Errorf("expected successful file path 'zoom.txt', but got, %v", s)
		}

		for _, fr := range got.FileResults {
			if fr.Err != nil {
				t.Errorf("expected no failed files but got, %v", fr)
			}
		}

	})

	t.Run("failing file in the system", func(t *testing.T) {
		fileSystem := errorFS{
			FS: fstest.MapFS{
				"zoom.txt": {
					Data: []byte(testHands),
				},
				"failure.txt": {
					Data: []byte("some data"),
				},
			},
			failOn: "failure.txt",
		}

		dir, _ := fs.ReadDir(fileSystem, ".")

		handsChannel := streamHands(fileSystem, dir)

		got := collectResults(handsChannel)

		for _, f := range got.FileResults {
			if !errors.Is(f.Err, ErrFileNotParsable) && f.Path == "failure.txt" {
				t.Errorf("wanted %#v error but got %#v", ErrFileNotParsable, f.Err)
			}

			if f.Path == "zoom.txt" && f.HandsParsed != 1 {
				t.Errorf("wanted 1 hands parsed but got %#v", f.HandsParsed)
			}
		}
	})
}

func TestExtractFileResults(t *testing.T) {

	t.Run("parsable files one successful path one unsuccessful path", func(t *testing.T) {
		const failureFileName string = "failures.txt"

		data := map[string]*fileCounter{
			"zoom.txt": {
				success: 121,
				failure: 1,
				err:     nil,
			},
			failureFileName: {
				success: 0,
				failure: 5,
				err:     nil,
			},
		}

		got := extractFileResults(data)
		want := []FileResult{
			{"zoom.txt", 121, 1, nil}, {failureFileName, 0, 5, ErrFailRate},
		}

		if len(got) != 2 {
			t.Errorf("expected length of fileResults to be 2 but got %v", len(got))
		}

		for _, g := range got {

			if g.Path == failureFileName && !errors.Is(g.Err, ErrFailRate) {
				t.Errorf("expected ErrFailRate for failures.txt but got %v", g.Err)
			}
		}

		if !errors.Is(want[1].Err, ErrFailRate) {
			t.Errorf("expected ErrFailRate but got %v", want[1].Err)
		}
	})

	t.Run("unparsable file and one parsable", func(t *testing.T) {
		const failureFileName string = "failures.txt"
		data := map[string]*fileCounter{
			"zoom.txt": {
				success: 121,
				failure: 1,
				err:     nil,
			},
			failureFileName: {
				success: 0,
				failure: 0,
				err:     ErrFileNotParsable,
			},
		}

		fileResults := extractFileResults(data)

		for _, f := range fileResults {
			if f.Err == nil && f.Path == failureFileName {
				t.Fatal("wanted an error but did't get one")
			}

			if !errors.Is(f.Err, ErrFileNotParsable) && f.Path == failureFileName {
				t.Errorf("wanted an ErrFileNotParsable but got %#v", f.Err)
			}
		}
	})

}

func TestCheckFailureRate(t *testing.T) {

	success := 101
	failure := 1
	ok := failRateExceeded(failure, success)

	if !ok {
		t.Fatal("wanted true checkFailureRate but was false")
	}
}

func sumHandsHelper(exportResult []FileResult) (successCount, failureCount int) {
	successCount = 0
	failureCount = 0
	for _, f := range exportResult {
		successCount += f.HandsParsed
		failureCount += f.HandErrs
	}
	return successCount, failureCount
}

type errorFS struct {
	fs.FS
	failOn string
}

func (e errorFS) Open(name string) (fs.File, error) {
	if name == e.failOn {
		return nil, fmt.Errorf("permission denied")
	}
	return e.FS.Open(name)
}
