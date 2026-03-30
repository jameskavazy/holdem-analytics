package pokerhud

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sync"
)

const maxFailRate float64 = 0.005

var (
	ErrFailRate        = errors.New("error handErrs exceeded the maximum fail rate")
	ErrFileNotParsable = errors.New("error failed to open file for parsing")
)

func FailRateErr(msg string) error {
	return fmt.Errorf("%w: %s", ErrFailRate, msg)
}

func FileNotParsableErr(msg string) error {
	return fmt.Errorf("%w: %s", ErrFileNotParsable, msg)
}

type ExportResult struct {
	FileResults []FileResult
	FsErr       error // filesystem error preventing any parsing
}

func (e *ExportResult) FileErrorCount() int {

	errCount := 0

	for _, f := range e.FileResults {
		if f.Err != nil {
			errCount++
		}
	}
	return errCount
}

func (e *ExportResult) HandsCount() int {
	count := 0
	for _, f := range e.FileResults {
		count += f.HandsParsed
	}
	return count
}

func (e *ExportResult) HandErrCount() int {
	count := 0
	for _, f := range e.FileResults {
		count += f.HandErrs
	}
	return count
}

func (e *ExportResult) SuccessCount() int {
	return len(e.FileResults) - e.FileErrorCount()
}

type fileCounter struct {
	success int
	failure int
	err     error
}

// ExportHands imports user hand history for the first time. Returns a slice of hands for insertion into the database.
func ExportHands(fileSystem fs.FS) ExportResult {
	dir, fsErr := fs.ReadDir(fileSystem, ".")

	if fsErr != nil {
		return ExportResult{
			nil,
			fsErr,
		}
	}

	// TODO create a worker pool if dir len > 10
	// runtime.NumCPU() for numWorkers range filesChannel and hands from sessionfile to hands CHan

	handsChannel := streamHands(fileSystem, dir)

	// map - add filename to map, and increment counter...
	return collectResults(handsChannel)
}

func streamHands(fileSystem fs.FS, dir []fs.DirEntry) <-chan handImport {
	var wg sync.WaitGroup
	handsChannel := make(chan handImport, 10000)

	for _, file := range dir {
		// TODO - move file once processed... also some sort of logic that works out once whole file is read to move it? Get Hands While Playing...
		// TODO - FILENAME will contain the currency type, set up some enums... etc.
		// Count the number of duplicates...
		if !file.IsDir() {
			wg.Add(1)
			go func(f fs.DirEntry) {
				defer wg.Done()
				ok, fsErr := extractHandsFromFile(fileSystem, f.Name(), handsChannel)

				if !ok {
					log.Printf("An error occurred parsing file %s: %#v", f.Name(), fsErr.Error())
					handsChannel <- handImport{filePath: f.Name(), fileErr: true}
				}
			}(file)
		}
	}

	go func() {
		wg.Wait()
		close(handsChannel)
	}()

	return handsChannel
}

func collectResults(handsChannel <-chan handImport) ExportResult {

	counter := map[string]*fileCounter{}

	for h := range handsChannel {

		// TODO We can receive the hands up to a 5k chunk and then commit to a database!
		if _, ok := counter[h.filePath]; !ok {
			counter[h.filePath] = &fileCounter{}
		}

		if h.fileErr {
			counter[h.filePath].err = FileNotParsableErr("could not open file")
		} else if h.handErr != nil {
			counter[h.filePath].failure++
		} else {
			counter[h.filePath].success++
		}
		// TODO: upon receiving the handImport we can pass off to our backend. Spawn another goroutine here?

	}
	fileResults := extractFileResults(counter)

	return ExportResult{
		fileResults,
		nil,
	}
}

func extractFileResults(results map[string]*fileCounter) []FileResult {
	fileResults := make([]FileResult, len(results))
	i := 0
	for k, v := range results {

		fr := FileResult{
			Path:        k,
			HandsParsed: v.success,
			HandErrs:    v.failure,
		}
		if v.err != nil {
			fr.Err = v.err
		} else if failRateExceeded(v.failure, v.success) {
			fr.Err = FailRateErr(fmt.Sprintf("%v successful, %v failed. Maximum fail rate: %v", v.success, v.failure, maxFailRate))
		}
		fileResults[i] = fr
		i++
	}
	return fileResults
}

func failRateExceeded(failure, success int) bool {
	total := float64(success) + float64(failure)
	if total == 0 {
		return false
	}
	return (float64(failure) / total) > maxFailRate
}
