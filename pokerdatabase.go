package pokerhud

// TODO - At some point we're going to want to make an interface of sorts so that we handle Pokerstars hands, Party poker hands. etc.
// so Hand will actually be an interface and there'll be a pokerstarshand struct that implements hand interface... methods TBD.
// equally each hand

// // We'd need some kind of continual running to scan the FS for new hand files and keep reading from the same particular one.
// func GetHandsWhilePlaying() {
//     TODO - detect latest file in HH fs. Open file & parse changes to it?
// }

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	handInfoDelimiter string = "\n\n\n"
	flopSignifier     string = "*** FLOP ***"
	turnSignifier     string = "*** TURN ***"
	riverSignifier    string = "*** RIVER ***" //TODO players can run it twice... >  *** FIRST RIVER *** *** SECOND RIVER ***
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
	ErrNoAction   = errors.New("error no action found on text line")
	ErrNoHandID   = errors.New("error no hand ID was found, unable parse. ignoring hand")
	ErrNoCurrency = errors.New("error parsing Action.Amount, expected currency'")
)

// CurrencyError formats sends an error accepting a msg to provide further information about causation
func CurrencyError(msg string) error {
	return fmt.Errorf("%w: %s", ErrNoCurrency, msg)
}

func NoHandIDError(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrNoHandID)
}

func NoActionError(msg string) error {
	return fmt.Errorf("%w: %s", ErrNoAction, msg)
}

// Hand represents a hand of poker
type Hand struct {
	ID             string
	Date           time.Time
	Players        []Player
	Actions        []Action
	CommunityCards []string
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
	Cards    string
}

type handImport struct {
	hands    []Hand
	handErrs []error
}

func (t ActionType) String() string {
	return string(t)
}

func (s *Street) Next(line string) {
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
		return nil, []error{errors.New("error reading file system")}
	}

	var allHands []Hand
	var handErrs []error

	allHandsChannel := make(chan handImport, len(dir))

	defer close(allHandsChannel)
	var wg sync.WaitGroup

	for _, file := range dir {
		wg.Add(1)
		// TODO - move file once processed... also some sort of logic that works out once whole file is read to move it? Get Hands While Playing...
		// Count handErrs so we can tell the user X amount of hands errors
		// Count the number of duplicates...

		go func(file fs.DirEntry) {
			defer wg.Done()
			sessionHands, sessionFileErr := handsFromSessionFile(fileSystem, file.Name())
			allHandsChannel <- handImport{sessionHands, sessionFileErr}
			//TODO move file.
		}(file)
	}

	wg.Wait()


	for i := 0; i < len(dir); i++ {
		h, _ := <-allHandsChannel

		if h.hands != nil {
			allHands = append(allHands, h.hands...)
			handErrs = append(handErrs, h.handErrs...)
		}
	}
	return allHands, handErrs
}

func handsFromSessionFile(filesystem fs.FS, filename string) ([]Hand, []error) {
	handData, err := fs.ReadFile(filesystem, filename)
	if err != nil {
		return nil, []error{err}
	}
	hands, parseErr := parseHandData(handData)
	return hands, parseErr
}

func parseHandData(fileData []byte) ([]Hand, []error) {
	sessionData := string(fileData)
	handsText := strings.Split(sessionData, handInfoDelimiter)

	var hands []Hand
	var errs []error

hands:
	for _, handText := range handsText {

		// One shot data - grab unique identifiers of hand
		handID := handIDFromText(handText)
		if handID == "" {
			shortHand := ellipsis(handText, 100)
			errs = append(errs, NoHandIDError(fmt.Sprintf("in hand %v", shortHand)))
			continue hands
		}
		dateTime := parseDateTime(dateTimeStringFromHandText(handText))

		// Loop through and append remaining data
		scanner := createHandScanner(handText)

		var players []Player
		var actions []Action
		var street = Preflop
		var order = 1
		var board []string = parseCommunityCards(handText)

		for scanner.Scan() {
			line := scanner.Text()

			actionResult, actionErr := ParseAndAppendActions(line, &street, actions, &order)
			if actionErr != nil {
				errs = append(errs, actionErr)
				continue hands
			}
			actions = actionResult

			if playersFound, found := playersFromText(line); found {
				players = append(players, playersFound)
			}
		}

		hands = append(hands, Hand{
			ID:             handID,
			Date:           dateTime,
			Players:        players,
			Actions:        actions,
			CommunityCards: board,
		})
	}
	return hands, errs
}

// ParseAndAppendActions builds an action from text data, and appends it to the existing Action slice before returning the now updated Action slice
func ParseAndAppendActions(line string, street *Street, actions []Action, order *int) ([]Action, error) {
	street.Next(line)
	updatedActions, err := parseAction(line, actions, street, order)
	if err != nil {
		if !errors.Is(err, ErrNoAction) || !errors.Is(err, ErrNoCurrency) {
			return updatedActions, err
		}
		return updatedActions, nil
	}
	return updatedActions, nil
}

func parseAction(line string, actions []Action, actionStreet *Street, order *int) ([]Action, error) {
	actionType, err := actionTypeFromText(line)

	if errors.Is(err, ErrNoAction) {
		return actions, nil
	}

	if err != nil {
		return actions, err
	}

	playerName, err := actionPlayerNameFromText(line)
	if err == nil {
		amount, _ := actionAmountFromText(line)

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
	(*order)++

	return actions, nil
}

// Returns a pointer to bufio.Scanner for parsing Hand data
func createHandScanner(h string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(h))
	return scanner
}

// Returns a hand Id string from the hand info string
func handIDFromText(handText string) string {
	if !strings.Contains(handText, "Hand #") {
		return ""
	}

	if strings.Contains(handText, ":") {
		return substringBetween(handText, "#", ":")
	}

	return ""
}

func actionTypeFromText(line string) (ActionType, error) {
	actionTypes := []ActionType{Posts, Folds, Checks, Bets, Calls, Raises}

	for _, t := range actionTypes {
		if strings.Contains(line, t.String()) {
			return t, nil
		}
	}
	return "", NoActionError(fmt.Sprintf("on line %v", line))
}

func actionPlayerNameFromText(line string) (string, error) {
	if strings.Contains(line, ":") {
		return strings.Split(line, ":")[0], nil
	}

	return "", errors.New("couldn't find player name to parse")
}

func actionAmountFromText(line string) (float64, error) {
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

	return 0, CurrencyError(fmt.Sprintf("on line %v", line))
}

func dateTimeStringFromHandText(line string) string {

	var timeString string
	if strings.ContainsAny(line, "[]") {
		timeString = substringBetween(line, "[", " ET]")
	}

	formattedTimeString := strings.Map(func(r rune) rune {
		if r == '/' {
			return '-'
		}
		return r
	}, timeString)
	return formattedTimeString
}

func parseDateTime(timeString string) time.Time {
	siteLocation, _ := time.LoadLocation("America/New_York")
	siteTime, _ := time.ParseInLocation(time.DateTime, timeString, siteLocation)
	return siteTime.Local()
}

func parseCommunityCards(handText string) []string {
	if strings.Contains(handText, "Hand was run twice") {
		firstBoard := parseBoardInfo(handText, "FIRST Board [", "]")
		secondBoard := parseBoardInfo(handText, "SECOND Board [", "]")
		return []string{firstBoard, secondBoard}
	}
	if strings.Contains(handText, "Board [") {
		board := parseBoardInfo(handText, "Board [", "]")
		return []string{board}
	}
	return nil
}

func parseBoardInfo(handText, boardStart, boardEnd string) string {
	return substringBetween(handText, boardStart, boardEnd)
}

// Substring between returns the substring between the first instance of characters start and end.
// If text does not contain the original text is returned unchanged
func substringBetween(text, start, end string) string {
	if !strings.Contains(text, start) || !strings.Contains(text, end) {
		return text
	}
	return strings.Split(strings.Split(text, start)[1], end)[0]
}

func playersFromText(line string) (Player, bool) {
	if strings.Contains(line, "showed [") {
		return parsePlayerInfo(line, "showed ["), true
	}

	if strings.Contains(line, "mucked [") {
		return parsePlayerInfo(line, "mucked ["), true
	}

	if strings.Contains(line, "folded") {
		return parsePlayerInfo(line, ""), true
	}

	return Player{}, false
}

func parsePlayerInfo(line string, cardPrefix string) Player {
	var playerName string
	var cards string

	if cardPrefix != "" {
		splitLine := strings.Split(line, cardPrefix)
		prefixWithPlayerName := splitLine[0]
		playerName = strings.Fields(prefixWithPlayerName)[2]
		cards = strings.Split(splitLine[1], "]")[0]
	} else {
		playerName = strings.Fields(line)[2]
		cards = ""
	}

	return Player{
		Username: playerName,
		Cards:    cards,
	}
}

// Ellipsis truncates a given string by the max length of characters provided and appends with ellipsis.
// If the length of string is less than the maxLen the whole string is return untruncated.
func ellipsis(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-3]) + "..."
}
