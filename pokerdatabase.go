package pokerhud

import (
	"bufio"
	"errors"
	"io/fs"
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

type Hand struct {
	Id        string
	Players   []string
	HeroCards string
	Actions   []Action
}

type Action struct {
	Player     string //TODO make player explicit type?
	Order      int
	Street     Street
	ActionType ActionType
	Amount     float64
}

type Street string

type ActionType string

func (t ActionType) String() string {
	return string(t)
}

// Todo - At some point we're going to want to make an interface of sorts so that we handle Pokerstars hands, Party poker hands. etc.
//so Hand will actually be an interface and there'll be a pokerstarshand struct that implements hand interface... methods TBD.
// equally each hand

// Do we need Action struct? Does that belong internally to the hand

// // We'd need some kind of continual running to scan the FS for new hand files and keep reading from the same particular one.
// func GetHandsWhilePlaying() {
//     //TODO - detect latest file in HH fs. Open file & parse changes to it?
// }

// Imports user hand history for the first time. Returns a slice of hands for insertion into the database.
func HandHistoryFromFS(fileSystem fs.FS) ([]Hand, error) {
	dir, err := fs.ReadDir(fileSystem, ".")

	if err != nil {
		return nil, errors.New("error reading filesystem")
	}

	var allHands []Hand

	for _, f := range dir {
		sessionHands := handsFromSessionFile(fileSystem, f.Name())
		allHands = slices.Concat(allHands, sessionHands)
	}

	return allHands, nil
}

func handsFromSessionFile(filesystem fs.FS, filename string) []Hand {
	handData, _ := fs.ReadFile(filesystem, filename)
	return parseHandData(handData)
}

func parseHandData(fileData []byte) []Hand {
	// TODO a custom type of HandRawData string might be beneficial here?
	sessionData := string(fileData)
	handsText := strings.Split(sessionData, handInfoDelimiter)

	var hands []Hand

	for _, h := range handsText {

		handId := handIdFromText(h)
		scanner := createHandScanner(h)

		var playerNames []string
		var actions []Action
		var heroCards string
		var street Street = Preflop
		var order int = 1

		// TODO does it make sense to scan the whole hand first, grab the players? and then start again this time with player info?
		for scanner.Scan() {
			playerNames = updatePlayerNames(scanner, playerNames)
			heroCards = setHeroCards(scanner, heroCards)
			actions = BuildActions(scanner, &street, actions, &order)
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
func BuildActions(scanner *bufio.Scanner, street *Street, actions []Action, order *int) []Action {
	getStreetFromText(scanner, street)
	return parseAction(scanner, actions, street, order)
}

func parseAction(scanner *bufio.Scanner, actions []Action, actionStreet *Street, order *int) []Action {
	actionType, err := actionTypeFromText(scanner)
	if err == nil {
		playerName, err := actionPlayerNameFromText(scanner)
		if err == nil {
			actions = append(actions, Action{
				ActionType: actionType,
				Player:     playerName,
				Street:     *actionStreet,
				Order:      *order,
				Amount:     actionAmountFromText(scanner),
			})
		}
		*order++
	}
	return actions
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
func updatePlayerNames(scanner *bufio.Scanner, playerNames []string) []string {
	nameFound := handplayerNameFromText(scanner)
	if nameFound != "" {
		playerNames = append(playerNames, nameFound)
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

func handplayerNameFromText(scanner *bufio.Scanner) string {
	var playerName string
	// Might need to pass in the street? because otherwise, there's Summary section that matches closely the same pattern
	//TODO Refactor this -> doesn't seem robust e.g. start + 2 seems like asking for a panic
	if strings.Contains(scanner.Text(), "Seat ") && strings.Contains(scanner.Text(), "chips") {
		start := strings.Index(scanner.Text(), ": ")
		end := strings.Index(scanner.Text(), " (")

		if start != -1 || end != -1 {
			playerName = scanner.Text()[start+2 : end]
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
	return "", errors.New("no action found on text line")

}

func actionPlayerNameFromText(scanner *bufio.Scanner) (string, error) {
	if strings.Contains(scanner.Text(), ":") {
		return strings.Split(scanner.Text(), ":")[0], nil
	}

	return "", errors.New("couldn't find player name to parse")
}

func actionAmountFromText(scanner *bufio.Scanner) float64 {
	if strings.Contains(scanner.Text(), Dollar) {
		amount, _ := strconv.ParseFloat((strings.Split(scanner.Text(), Dollar)[1]), 64)
		return amount
	}
	return 0
}
