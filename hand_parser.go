package pokerhud

import (
	"bufio"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Delimiter and signifier constants for parsing hand files
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

func parseHands(fileData []byte) ([]Hand, []error) {
	sessionData := string(fileData)
	handsText := strings.Split(sessionData, handInfoDelimiter)

	var hands []Hand
	var errs []error

	for _, handText := range handsText {
		handID, dateTime, board, infoErr := parseHandSummary(handText)
		if infoErr != nil {
			errs = append(errs, infoErr)
			continue // the hand lacks crucial metadata - skip
		}

		players, actions, actionErr := parseActions(handText)
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

func parseHandSummary(handText string) (handID string, dateTime time.Time, board []string, err error) {
	handID = handIDFromText(handText)
	if handID == "" {
		shortHand := ellipsis(handText, 100) // Truncate hand info for error
		err = NoHandIDError(fmt.Sprintf("in hand %v", shortHand))
	}
	dateTime = parseDateTime(dateTimeFromText(handText))
	board = parseCommunityCards(handText)
	return
}

func parseActions(handText string) ([]Player, []Action, error) {

	var players []Player
	var actions []Action
	var street = Preflop
	var order = 0

	scanner := createHandScanner(handText)

	for scanner.Scan() {
		line := scanner.Text()

		actionResult, actionFound, actionErr := parseActionLine(line, &street, &order)

		if actionErr != nil {
			return nil, nil, actionErr
		}

		if actionFound {
			actions = append(actions, actionResult)
		}

		if player, ok := parsePlayer(line); ok {
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

// parseActionLine checks a line of text for a poker action and if found returns an action, along
// with true bool and nil error. If there is no action found, an empty Action struct will be returned,
// along with a false bool. If there was an error parsing an action detail a non-nil error will be returned.
func parseActionLine(line string, actionStreet *Street, order *int) (Action, bool, error) {
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

func parseCommunityCards(handText string) []string {
	if strings.Contains(handText, "Hand was run twice") {
		firstBoard := communityCardsFromText(handText, ritFirstBoardSignifier, "]")
		secondBoard := communityCardsFromText(handText, ritSecondBoardSignifier, "]")
		return []string{firstBoard, secondBoard}
	}
	if strings.Contains(handText, boardSignifier) {
		board := communityCardsFromText(handText, boardSignifier, "]")
		return []string{board}
	}
	return nil
}

func communityCardsFromText(handText, boardStart, boardEnd string) string {
	return substringBetween(handText, boardStart, boardEnd)
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


func parsePlayer(line string) (Player, bool) {
	if strings.Contains(line, "Dealt to") {
		return playerInfoFromText(line, "["), true
	}

	if strings.Contains(line, showedSignifier) {
		return playerInfoFromText(line, showedSignifier), true
	}

	if strings.Contains(line, muckedSignifier) {
		return playerInfoFromText(line, muckedSignifier), true
	}

	if strings.Contains(line, foldedSignifier) {
		return playerInfoFromText(line, ""), true
	}

	if strings.Contains(line, collectedSignifier) {
		return playerInfoFromText(line, ""), true
	}

	return Player{}, false
}

func playerInfoFromText(line string, cardPrefix string) Player {
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

func parseDateTime(timeString string) time.Time {
	siteLocation, _ := time.LoadLocation("America/New_York")
	siteTime, _ := time.ParseInLocation(time.DateTime, timeString, siteLocation)
	return siteTime.Local()
}

// dateTimeFromText extracts the relevant time from the hand information
// and converts to a string that can be transformed into a time.Time
func dateTimeFromText(line string) string {

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


// Substring between returns the substring between the first instance of characters start and end.
// If text does not contain the original text is returned unchanged
func substringBetween(text, start, end string) string {
	if !strings.Contains(text, start) || !strings.Contains(text, end) {
		return text
	}
	return strings.Split(strings.Split(text, start)[1], end)[0]
}

// ellipsis truncates a given string by the max length of characters provided and appends with ellipsis.
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

// CreateHandScanner returns a pointer to bufio.Scanner for parsing Hand data
func createHandScanner(h string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(h))
	return scanner
}