package hands

// TODO - At some point we're going to want to make an interface of sorts so that we handle Pokerstars hands, Party poker hands. etc.
// so Hand will actually be an interface and there'll be a pokerstarshand struct that implements hand interface... methods TBD.
// equally each hand

// // We'd need some kind of continual running to scan the FS for new hand files and keep reading from the same particular one.
// func GetHandsWhilePlaying() {
//     TODO - detect latest file in HH fs. Open file & parse changes to it?
// }

// TODO - RETURN uncalled bet needs to be parsed otherwise the valulations of what happened just aren't right...

import (
	"errors"
	"fmt"
	"strings"
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

// Global Errs
var (
	ErrFailToParseAction = errors.New("error no action found on text line")
	ErrNoHandID          = errors.New("error no hand ID was found, unable parse. ignoring hand") // TODO: Perhaps make this a struct & add the file, hand info and error to struct.
	ErrPlayerInfo        = errors.New("error could not parse player info, not enough fields on line. hand data is corrupt")
	errNoCurrency        = errors.New("error parsing Action.Amount, expected currency'")
)

// CurrencyError propagate an errNoCurrency error with customised message msg.
func CurrencyError(msg string) error {
	return fmt.Errorf("%w: %s", errNoCurrency, msg)
}

// NoHandIDError  propagate an ErrNoHandID error with customised message msg.
func NoHandIDError(msg string) error {
	return fmt.Errorf("%w: %s", ErrNoHandID, msg)
}

// ActionParseError propagate an ErrFailToParseAction error with customised message msg.
func ActionParseError(msg string) error {
	return fmt.Errorf("%w: %s", ErrFailToParseAction, msg)
}

func PlayerInfoError(msg string) error {
	return fmt.Errorf("%w: %s", ErrPlayerInfo, msg)
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
	CommunityCards []string //[][]Card
	Pot            float64
	Rake           float64
	// UncalledBet    float64
	// Winners        []Winner
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
	Cards    string //[]Card
	// Seat      int
	// ChipCount float64
}

type Card string

// type Winner struct {
// 	PlayerName string
// 	Amount     float64
// }

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
