package pokerhud

// TODO - At some point we're going to want to make an interface of sorts so that we handle Pokerstars hands, Party poker hands. etc.
// so Hand will actually be an interface and there'll be a pokerstarshand struct that implements hand interface... methods TBD.
// equally each hand

// // We'd need some kind of continual running to scan the FS for new hand files and keep reading from the same particular one.
// func GetHandsWhilePlaying() {
//     //TODO - detect latest file in HH fs. Open file & parse changes to it?
// }

// TODO - Add waitgroups, goroutines and chans for concurrent file parsing...
//

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"slices"
	"strconv"
	"strings"
)

const (
	handInfoDelimiter string = "\n\n\n"
	flopSignifier     string = "*** FLOP ***"
	turnSignifier     string = "*** TURN ***"
	riverSignifier    string = "*** RIVER ***"
	Dollar            string = "$"

	Preflop Street = "preflop"
	Flop    Street = "flop"
	Turn    Street = "turn"
	River   Street = "river"

	Folds  ActionType = "folds"
	Checks ActionType = "checks"
	Calls  ActionType = "calls"
	Bets   ActionType = "bets"
	Raises ActionType = "raises"
	Posts  ActionType = "posts"
)

var (
	ErrNoAction   = errors.New("no action found on text line")
	ErrNoCurrency = errors.New("error parsing Action.Amount, line expected a monetary amount not contain '$'")
)

// Hand represents a hand of poker
type Hand struct {
	Id        string
	Players   []Player
	HeroCards string
	Actions   []Action
}

// Action is a representation of individual actions made by players within a specific hand
type Action struct {
	Player     Player // TODO make player explicit type?
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
}

func (t ActionType) String() string {
	return string(t)
}

// Imports user hand history for the first time. Returns a slice of hands for insertion into the database.
func HandHistoryFromFS(fileSystem fs.FS) ([]Hand, error) {
	dir, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return nil, errors.New("error reading file system")
	}

	var allHands []Hand

	for _, f := range dir {
		sessionHands, err := handsFromSessionFile(fileSystem, f.Name())
		if err != nil {
			return nil, fmt.Errorf("error reading file: %v", f.Name())
		}

		allHands = slices.Concat(allHands, sessionHands)
	}

	return allHands, nil
}

func handsFromSessionFile(filesystem fs.FS, filename string) ([]Hand, error) {
	handData, err := fs.ReadFile(filesystem, filename)
	if err != nil {
		return nil, err
	}
	return parseHandData(handData), nil
}

func parseHandData(fileData []byte) []Hand {
	sessionData := string(fileData)
	handsText := strings.Split(sessionData, handInfoDelimiter)

	var hands []Hand

	for _, h := range handsText {

		handId := handIdFromText(h)
		scanner := createHandScanner(h)

		var playerNames []Player
		var actions []Action
		var heroCards string
		var street Street = Preflop
		var order int = 1

		for scanner.Scan() {
			playerNames = updatePlayerNames(scanner, playerNames)
			heroCards = setHeroCards(scanner, heroCards)
			actionResult, actionErr := ParseAndAppendActions(scanner, &street, actions, &order)
			if actionErr != nil {
				log.Println(actionErr)
			}
			actions = actionResult
		}

		hands = append(hands, Hand{
			Id:        handId,
			Players:   playerNames,
			HeroCards: heroCards,
			Actions:   actions,
		})
	}
	return hands
}

// Builds action from text data, appends to the existing slice and returns back an updated slice
func ParseAndAppendActions(scanner *bufio.Scanner, street *Street, actions []Action, order *int) ([]Action, error) {
	getStreetFromText(scanner, street)
	updatedActions, err := parseAction(scanner, actions, street, order)
	if err != nil {
		if err != ErrNoAction || err != ErrNoCurrency {
			return updatedActions, err
		}
		return updatedActions, nil
	}
	return updatedActions, nil
}

func parseAction(scanner *bufio.Scanner, actions []Action, actionStreet *Street, order *int) ([]Action, error) {
	actionType, err := actionTypeFromText(scanner)
	if err == ErrNoAction {
		return actions, nil
	}

	if err != nil {
		return actions, err
	}

	playerName, err := actionPlayerNameFromText(scanner)
	if err == nil {
		amount, amountErr := actionAmountFromText(scanner)
		if amountErr != nil {
			log.Println(amountErr)
		}
		actions = append(actions, Action{
			ActionType: actionType,
			Player: Player{
				Username: playerName,
			},
			Street: *actionStreet,
			Order:  *order,
			Amount: amount,
		})
	}
	*order++

	return actions, nil
}

func getStreetFromText(scanner *bufio.Scanner, actionStreet *Street) {
	switch {
	case strings.Contains(scanner.Text(), flopSignifier):
		*actionStreet = Flop
	case strings.Contains(scanner.Text(), turnSignifier):
		*actionStreet = Turn
	case strings.Contains(scanner.Text(), riverSignifier):
		*actionStreet = River
	}
}

func setHeroCards(scanner *bufio.Scanner, heroCards string) string {
	if heroCardsFromText(scanner) != "" {
		heroCards = heroCardsFromText(scanner)
	}
	return heroCards
}

// Extracts player name and updates playerNames slice for the Hand. If unable to extract a playername, the original playerNames slice is returned.
func updatePlayerNames(scanner *bufio.Scanner, playerNames []Player) []Player {
	nameFound := handPlayerNameFromText(scanner)
	if nameFound != "" {
		playerNames = append(playerNames, Player{
			Username: nameFound,
		})
	}
	return playerNames
}

// Returns a pointer to bufio.Scanner for parsing Hand data
func createHandScanner(h string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(h))
	return scanner
}

// Returns a hand Id string from the hand info string
func handIdFromText(h string) string {
	return strings.Split(strings.Split(h, ":")[0], "#")[1]
}

func handPlayerNameFromText(scanner *bufio.Scanner) string {
	var playerName string
	// Might need to pass in the street? because otherwise, there's Summary section that matches closely the same pattern
	// TODO Refactor this -> doesn't seem robust e.g. start + 2 seems like asking for a panic
	if strings.Contains(scanner.Text(), "Seat ") && strings.Contains(scanner.Text(), "chips") {
		start := strings.Index(scanner.Text(), ": ")
		end := strings.Index(scanner.Text(), " (")

		if start != -1 || end != -1 {
			playerName = (scanner.Text()[start+2 : end])
		}
	}
	return playerName
}

func heroCardsFromText(scanner *bufio.Scanner) string {
	var heroCards string

	if strings.Contains(scanner.Text(), "Dealt to") {
		start := strings.Index(scanner.Text(), "[")
		end := strings.Index(scanner.Text(), "]")
		if start != -1 && end != -1 {
			heroCards = scanner.Text()[start+1 : end]
			return heroCards
		}
	}

	return heroCards
}

func actionTypeFromText(scanner *bufio.Scanner) (ActionType, error) {
	actionTypes := []ActionType{Posts, Folds, Checks, Bets, Calls, Raises}

	for _, t := range actionTypes {
		if strings.Contains(scanner.Text(), t.String()) {
			return t, nil
		}
	}
	return "", ErrNoAction
}

func actionPlayerNameFromText(scanner *bufio.Scanner) (string, error) {
	if strings.Contains(scanner.Text(), ":") {
		return strings.Split(scanner.Text(), ":")[0], nil
	}

	return "", errors.New("couldn't find player name to parse")
}

func actionAmountFromText(scanner *bufio.Scanner) (float64, error) {
	// Strings.SplitN????
	line := scanner.Text()

	if strings.Contains(line, Dollar) && !strings.Contains(line, " to ") {
		amount, err := strconv.ParseFloat((strings.Split(line, Dollar)[1]), 64)
		if err != nil {
			return amount, fmt.Errorf("received %v parsing line %v", err, line)
		}
		return amount, nil
	}

	if strings.Contains(line, Dollar) && strings.Contains(line, " to ") {
		raiseAmt, _ := strings.CutSuffix(strings.Split(line, Dollar)[1], " to ")
		amount, err := strconv.ParseFloat(raiseAmt, 64)
		if err != nil {
			return amount, fmt.Errorf("received %v parsing line %v", err, line)
		}
		return amount, nil
	}

	if strings.Contains(line, "checks") || strings.Contains(line, "folds") {
		return 0, nil
	}

	ErrNoCurrency = fmt.Errorf("error on line %v parsing Action.Amount, line expected a monetary amount not contain '$'", line)
	return 0, ErrNoCurrency
}
