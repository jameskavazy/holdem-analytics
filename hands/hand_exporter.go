package hands

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sync"
)

const maxFailRate float64 = 0.005

var (
	// ErrFailRate indicates that the number of hand errors within the file exceeds the maximum fail rate constant
	ErrFailRate = errors.New("error handErrs exceeded the maximum fail rate")

	// ErrFileNotParsable indicates that the given file was not able to be opened or read by the scanner
	ErrFileNotParsable = errors.New("error failed to open file for parsing")
)

// FailRateErr returns an error containing an ErrFailRate and message
func FailRateErr(msg string) error {
	return fmt.Errorf("%w: %s", ErrFailRate, msg)
}

// FileNotParsableErr returns an error containing an ErrFileNotParsable and message
func FileNotParsableErr(msg string) error {
	return fmt.Errorf("%w: %s", ErrFileNotParsable, msg)
}

// ExportResult contains a slice of FileResult stats and an FsErr that reports on any filesystem errors encountered during the export
type ExportResult struct {
	FileResults []FileResult
	FsErr       error // filesystem error preventing any parsing
}

// FileResult provides information about the file parsed, including path, number of successful hands/hand errors.
type FileResult struct {
	Path        string
	HandsParsed int
	HandErrs    int
	Err         error
}

// FileErrorCount returns the number of files in the ExportResult with a non-nil file error.
func (e *ExportResult) FileErrorCount() int {
	errCount := 0

	for _, f := range e.FileResults {
		if f.Err != nil {
			errCount++
		}
	}
	return errCount
}

// HandsCount returns the number of successfully parsed hands across all files within the ExportResult.
func (e *ExportResult) HandsCount() int {
	count := 0
	for _, f := range e.FileResults {
		count += f.HandsParsed
	}
	return count
}

// HandErrCount returns the number of hands across all files within the ExportResult that failed to be parsed.
func (e *ExportResult) HandErrCount() int {
	count := 0
	for _, f := range e.FileResults {
		count += f.HandErrs
	}
	return count
}

// SuccessCount returns the number of files in the ExportResult that were successfully parsed with no file errors.
func (e *ExportResult) SuccessCount() int {
	return len(e.FileResults) - e.FileErrorCount()
}

// SuccessFiles returns a list of file names that were successful
func (e *ExportResult) SuccessFiles() []string {
	successFiles := make([]string, 0, 3)
	for _, s := range e.FileResults {
		if s.Err == nil {
			successFiles = append(successFiles, s.Path)
		}
	}
	return successFiles
}

type handImport struct {
	filePath string
	hand     Hand
	handErr  error
	fileErr  bool
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

		if _, ok := counter[h.filePath]; !ok {
			counter[h.filePath] = &fileCounter{}
		}

		if h.fileErr {
			counter[h.filePath].err = FileNotParsableErr("could not open file")
		} else if h.handErr != nil {
			counter[h.filePath].failure++

			handID := h.hand.Metadata.ID
			if handID == "" {
				handID = "no hand id"
			}
			log.Printf("got an error parsing hand %v in %v: %v", handID, h.filePath, h.handErr.Error())
		} else {
			counter[h.filePath].success++
		}
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
