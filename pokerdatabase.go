package pokerhud

import (
	"bufio"
	"errors"
	"io/fs"
	"slices"
	"strings"
)

const handInfoDelimiter string = "\n\n\n"

type Hand struct {
	Id        string
	Players   []string
	HeroCards string
	Actions   []Action
}

type Action struct {
	// Player string //TODO make player explicit type?
	// Order  int
	// Street string
	ActionType string
	// Amount float64
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
	dir, _ := fs.ReadDir(fileSystem, ".")

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

		// TODO does it make sense to scan the whole hand first, grab the players? and then start again this time with player info?
		for scanner.Scan() {
			playerNames = updatePlayerNames(scanner, playerNames)
			heroCards = setHeroCards(scanner, heroCards)
			actions = BuildActions(scanner, actions)
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

func BuildActions(scanner *bufio.Scanner, actions []Action) []Action {
	actionType, err := actionTypeFromText(scanner)
	if err == nil {
		actions = append(actions, Action{
			ActionType: actionType,
		})
	}
	return actions
}

func setHeroCards(scanner *bufio.Scanner, heroCards string) string {
	if heroCardsFromText(scanner) != "" {
		heroCards = heroCardsFromText(scanner)
	}
	return heroCards
}

// Extracts player name and updates playerNames slice for the Hand. If unable to extract a playername, the original playerNames slice is returned.
func updatePlayerNames(scanner *bufio.Scanner, playerNames []string) []string {
	nameFound := playerNameFromText(scanner)
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

func playerNameFromText(scanner *bufio.Scanner) string {
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

func actionTypeFromText(scanner *bufio.Scanner) (string, error) {

	actionTypes := []string{
		"posts",
		"bets",
		"calls",
		"folds",
		"raises",
		"checks",
	}

	for _, t := range actionTypes {
		if strings.Contains(scanner.Text(), t) {
			return t, nil

		}
	}
	return "", errors.New("no action found on text line")

	// var order int = 1
	// var street string

	// switch {
	// case strings.Contains(scanner.Text(), "*** FLOP ***"):
	// 	street = "flop"
	// case strings.Contains(scanner.Text(), "*** TURN ***"):
	// 	street = "turn"
	// case strings.Contains(scanner.Text(), "*** RIVER ***"):
	// 	street = "river"
	// default:
	// 	street = "preflop"
	// }

	// scannedFieldData := strings.Fields(scanner.Text())

	// fmt.Println("actioinsFromText data before internal loop", scannedFieldData)
	// var playerName string
	// var actionType string
	// var amount float64

	// for _, d := range scannedFieldData {

	// 	if strings.Contains(d, playerNameFromText(scanner)) {
	// 		playerName, _ = strings.CutSuffix(d, ":")
	// 		// fmt.Println("actionsFromText playerName found ", playerName)
	// 	}

	// 	// switch {
	// 	// case strings.Contains(d, "posts"):
	// 	// 	actionType = "post"
	// 	// case strings.Contains(d, "bets"):
	// 	// 	actionType = "bet"
	// 	// case strings.Contains(d, "calls"):
	// 	// 	actionType = "call"
	// 	// case strings.Contains(d, "folds"):
	// 	// 	actionType = "fold"
	// 	// 	amount = 0
	// 	// case strings.Contains(d, "raises"):
	// 	// 	actionType = "raise"
	// 	// case strings.Contains(d, "checks"):
	// 	// 	actionType = "checks"
	// 	// 	amount = 0
	// 	// }

	// 	// if strings.Contains(d, "$") { // TODO const DOLLAR , if currency = dollar
	// 	// 	amount, _ = strconv.ParseFloat((strings.Split(d, "$")[1]), 64)
	// 	// }

	// }

	// order++

}
