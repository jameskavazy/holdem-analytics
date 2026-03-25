package pokerhud

// TODO - At some point we're going to want to make an interface of sorts so that we handle Pokerstars hands, Party poker hands. etc.
// so Hand will actually be an interface and there'll be a pokerstarshand struct that implements hand interface... methods TBD.
// equally each hand

// // We'd need some kind of continual running to scan the FS for new hand files and keep reading from the same particular one.
// func GetHandsWhilePlaying() {
//     TODO - detect latest file in HH fs. Open file & parse changes to it?
// }

// TODO - RETURN uncalled bet needs to be parsed otherwise the valulations of what happened just aren't right...

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Street represents stages of a poker hand
const (
	Preflop Street = "preflop"
	Flop    Street = "flop"
	Turn    Street = "turn"
	River   Street = "river"
)

// ActionType - the type of player actions
const (
	Folds  ActionType = "folds"
	Checks ActionType = "checks"
	Calls  ActionType = "calls"
	Bets   ActionType = "bets"
	Raises ActionType = "raises"
	Posts  ActionType = "posts"
)

// Currencies constants
const (
	Dollar string = "$"
)

var wg sync.WaitGroup

// Global Errs
var (
	ErrFailToParseAction = errors.New("error no action found on text line")
	ErrNoHandID          = errors.New("error no hand ID was found, unable parse. ignoring hand") // TODO: Perhaps make this a struct & add the file, hand info and error to struct.
	errNoCurrency        = errors.New("error parsing Action.Amount, expected currency'")
)

// CurrencyError propagate an errNoCurrency error with customised message msg.
func CurrencyError(msg string) error {
	return fmt.Errorf("%w: %s", errNoCurrency, msg)
}

// NoHandIDError  propagate an ErrNoHandID error with customised message msg.
func NoHandIDError(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrNoHandID)
}

// ActionParseError propagate an ErrFailToParseAction error with customised message msg.
func ActionParseError(msg string) error {
	return fmt.Errorf("%w: %s", ErrFailToParseAction, msg)
}

// Hand represents a hand of poker
type Hand struct {
	Metadata Metadata
	Players  []Player
	Actions  []Action
	Summary  Summary
}

type Metadata struct {
	ID   string
	Date time.Time
}

type Summary struct {
	CommunityCards []string
	Pot            float64
	Rake           float64
}

// Action is a representation of individual actions made by players within a specific hand
type Action struct {
	PlayerName string
	Order      int
	Street     Street
	ActionType ActionType
	Amount     float64
}

// Street is a string representation of the poker street an action was made on
type Street string

// ActionType is a the type of action made by a player. E.g. folds
type ActionType string

// Player - a player in the hand
type Player struct {
	Username string
	Cards    string
}

type handImport struct {
	hand    Hand
	handErr error
}

func (t ActionType) String() string {
	return string(t)
}

func (s *Street) next(line string) {
	switch {
	case strings.Contains(line, flopSignifier):
		*s = Flop
	case strings.Contains(line, turnSignifier):
		*s = Turn
	case strings.Contains(line, riverSignifier):
		*s = River
	}
}

// HandHistoryFromFS imports user hand history for the first time. Returns a slice of hands for insertion into the database.
func HandHistoryFromFS(fileSystem fs.FS) ([]Hand, []error) {
	dir, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return nil, []error{err} //errors.New("error reading file system")
	}

	var allHands []Hand
	var handErrs []error

	handsChannel := make(chan handImport, 10000)

	// TODO create a worker pool if dir len > 10

	for _, file := range dir {
		// TODO - move file once processed... also some sort of logic that works out once whole file is read to move it? Get Hands While Playing...
			// TODO - FILENAME will contain the currency type, set up some enums... etc.
		// Count handErrs so we can tell the user X amount of hands errors
		// Count the number of duplicates...
		if file.IsDir() {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, fsErr := handsFromSessionFile(fileSystem, file.Name(), handsChannel)

			if !ok {
				log.Fatal("An error occured parsing file", fsErr)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(handsChannel)
	}()

	for h := range handsChannel {

		// TODO In for range dir we can receive the hands up to a 5k chunk and then commit to a database!

		if !reflect.DeepEqual(h.hand, Hand{}) {
			allHands = append(allHands, h.hand)
			// TODO: upon receiving the handImport we can pass off to our backend. Spawn another goroutine here?
		}
		if h.handErr != nil {
			handErrs = append(handErrs, h.handErr)
		}
	}

	return allHands, handErrs
}

func handsFromSessionFile(filesystem fs.FS, filename string, handChan chan<- handImport) (ok bool, fsErr error) {
	file, err := filesystem.Open(filename)

	if err != nil {
		return false, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	result, scanErr := parseHands(scanner, handChan)

	if !result {
		return false, scanErr
	}

	return true, nil
}
