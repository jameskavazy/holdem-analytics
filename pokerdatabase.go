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
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// handInfoDelimiter  string = "\r\n\r\n\r\n"
	handInfoDelimiter       string = "\n\n\n"
	flopSignifier           string = "*** FLOP ***"
	turnSignifier           string = "*** TURN ***"
	riverSignifier          string = "*** RIVER ***" //TODO players can run it twice... >  *** FIRST RIVER *** *** SECOND RIVER ***
	showedSignifier         string = "showed ["
	muckedSignifier         string = "mucked ["
	foldedSignifier         string = "folded"
	collectedSignifier      string = "collected ("
	boardSignifier          string = "Board ["
	ritFirstBoardSignifier  string = "FIRST Board ["
	ritSecondBoardSignifier string = "SECOND Board ["
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
	ID             string
	Date           time.Time
	Players        []Player
	Actions        []Action
	CommunityCards []string
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
	hands    []Hand
	handErrs []error
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

	allHandsChannel := make(chan handImport, len(dir))

	defer close(allHandsChannel)

	for _, file := range dir {
		// TODO - move file once processed... also some sort of logic that works out once whole file is read to move it? Get Hands While Playing...
		// Count handErrs so we can tell the user X amount of hands errors
		// Count the number of duplicates...

		go func() {
			sessionHands, sessionFileErr := handsFromSessionFile(fileSystem, file.Name())
			allHandsChannel <- handImport{sessionHands, sessionFileErr}
		}()
	}

	for i := 0; i < len(dir); i++ {
		h := <-allHandsChannel

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
	hands, parseErrs := parseHandData(handData)
	return hands, parseErrs
}

func parseHandData(fileData []byte) ([]Hand, []error) {
	sessionData := string(fileData)
	handsText := strings.Split(sessionData, handInfoDelimiter)

	var hands []Hand
	var errs []error

	for _, handText := range handsText {
		handID, dateTime, board, infoErr := summaryInfoFromText(handText)
		if infoErr != nil {
			errs = append(errs, infoErr)
			continue // the hand lacks crucial metadata - skip
		}

		players, actions, actionErr := actionsFromText(handText)
		if actionErr != nil {
			errs = append(errs, actionErr)
			continue // the hand lacks crucial gameplay info - skip
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

// ParseAction checks a line of text for a poker action and if found, creates and appends it to the existing list.
// If there is no action found, the original Action slice will be returned. If there was an error parsing an action detail, the actions slice will become nil and void along with a non-nil error will be returned.
func parseAction(line string, actionStreet *Street, order *int) (Action, bool, error) {
	actionStreet.next(line)
	actionType, actionFound := actionTypeFromText(line)

	if !actionFound {
		return Action{}, actionFound, nil
	}

	// TODO - add player struct to the action rather than just the player name?
	playerName, playerErr := actionPlayerNameFromText(line)
	amount, amtErr := actionAmountFromText(line)

	// TODO - return a more specific error
	if playerErr != nil {
		return Action{}, actionFound, ActionParseError(fmt.Sprintf("%v %v", playerErr, line))
	}

	if amtErr != nil {
		return Action{}, actionFound, ActionParseError(fmt.Sprintf("%v %v", amtErr, line))
	}

	*order++

	return Action{
		ActionType: actionType,
		PlayerName: playerName,
		Street:     *actionStreet,
		Order:      *order,
		Amount:     amount,
	}, actionFound, nil

}

// CreateHandScanner returns a pointer to bufio.Scanner for parsing Hand data
func createHandScanner(h string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(h))
	return scanner
}

// handIDFromText returns the hand ID string from the hand info string
func handIDFromText(handText string) string {
	if !strings.Contains(handText, "Hand #") {
		return ""
	}

	if strings.Contains(handText, ":") {
		return substringBetween(handText, "#", ":")
	}

	return ""
}

func actionTypeFromText(line string) (ActionType, bool) {
	actionTypes := []ActionType{Posts, Folds, Checks, Bets, Calls, Raises}

	for _, t := range actionTypes {
		if strings.Contains(line, t.String()) {
			return t, true
		}
	}
	return "", false
}

func actionPlayerNameFromText(line string) (string, error) {
	if strings.Contains(line, ":") {
		return strings.Split(line, ":")[0], nil
	}
	return "", errors.New("could not parse name in action")
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
		firstBoard := parseBoardInfo(handText, ritFirstBoardSignifier, "]")
		secondBoard := parseBoardInfo(handText, ritSecondBoardSignifier, "]")
		return []string{firstBoard, secondBoard}
	}
	if strings.Contains(handText, boardSignifier) {
		board := parseBoardInfo(handText, boardSignifier, "]")
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

func playerFromText(line string) (Player, bool) {
	if strings.Contains(line, "Dealt to") {
		return parsePlayerInfo(line, "["), true
	}

	if strings.Contains(line, showedSignifier) {
		return parsePlayerInfo(line, showedSignifier), true
	}

	if strings.Contains(line, muckedSignifier) {
		return parsePlayerInfo(line, muckedSignifier), true
	}

	if strings.Contains(line, foldedSignifier) {
		return parsePlayerInfo(line, ""), true
	}

	if strings.Contains(line, collectedSignifier) {
		return parsePlayerInfo(line, ""), true
	}

	return Player{}, false
}

func parsePlayerInfo(line string, cardPrefix string) Player {
	playerName := strings.Fields(line)[2]
	var cards string

	if cardPrefix != "" {
		cards = substringBetween(line, cardPrefix, "]")
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

func summaryInfoFromText(handText string) (handID string, dateTime time.Time, board []string, err error) {
	handID = handIDFromText(handText)
	if handID == "" {
		shortHand := ellipsis(handText, 100) // Truncate hand info for error
		err = NoHandIDError(fmt.Sprintf("in hand %v", shortHand))
	}
	dateTime = parseDateTime(dateTimeStringFromHandText(handText))
	board = parseCommunityCards(handText)
	return
}

func actionsFromText(handText string) ([]Player, []Action, error) {

	var players []Player
	var actions []Action
	var street = Preflop
	var order = 0

	scanner := createHandScanner(handText)

	for scanner.Scan() {
		line := scanner.Text()

		actionResult, actionFound, actionErr := parseAction(line, &street, &order)

		if actionErr != nil {
			return nil, nil, actionErr
		}

		if actionFound {
			actions = append(actions, actionResult)
		}

		if player, ok := playerFromText(line); ok {
			if !slices.ContainsFunc(players, func(p Player) bool {
				return p.Username == player.Username
			}) {
				// Hero player is parsed before summary, so skip to avoid duplicate
				players = append(players, player)
			}
		}
	}
	return players, actions, nil
}
