package hands

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Delimiter and signifier constants for parsing hand files
const (
	handInfoDelimiter       string = "\nPokerStars "
	newLine                 string = "\n"
	flopSignifier           string = "*** FLOP ***"
	turnSignifier           string = "*** TURN ***"
	riverSignifier          string = "*** RIVER ***"
	summarySignifier        string = "*** SUMMARY ***"
	heroHandPrefix          string = "Dealt to"
	showedSignifier         string = "showed ["
	muckedSignifier         string = "mucked ["
	foldedSignifier         string = "folded"
	collectedSignifier      string = "collected ("
	boardSignifier          string = "Board ["
	ritFirstBoardSignifier  string = "FIRST Board ["
	ritSecondBoardSignifier string = "SECOND Board ["
	potSizeSignifier        string = "Total pot "
	rakeSizeSignifier       string = "Rake "
	// TODO: Uncalled bet ($0.04) returned to Folding3bets
)

var amountRegex = regexp.MustCompile(`\$(\d+(?:\.\d+)?)`)
var seatIntRegex = regexp.MustCompile(`^Seat (\d+): `)
var siteLocation, _ = time.LoadLocation("America/New_York")

func extractHandsFromFile(filesystem fs.FS, filename string, handChan chan<- handImport) (ok bool, fsErr error) {
	file, err := filesystem.Open(filename)

	if err != nil {
		return false, err
	}

	defer func() {
		closeErr := file.Close()

		if err == nil {
			err = closeErr
		}

	}()

	scanner := bufio.NewScanner(file)

	result, scanErr := parseHands(filename, scanner, handChan)

	if !result {
		return false, scanErr
	}

	return true, nil
}

func parseHands(filename string, fileData *bufio.Scanner, handChan chan<- handImport) (ok bool, scanErr error) {
	fileData.Split(splitByHands())

	for fileData.Scan() {
		handText := fileData.Text()

		metadata, metadataErr := parseMetaData(handText)

		if metadataErr != nil {
			handChan <- handImport{filename, Hand{}, metadataErr, false}
			continue // the hand lacks crucial metadata - skip
		}

		players, actions, winners, scanHandErr := scanHandLines(handText)
		if scanHandErr != nil {
			handChan <- handImport{
				filePath: filename,
				hand:     Hand{},
				handErr:  scanHandErr,
				fileErr:  false,
			}
			continue // the hand lacks crucial gameplay info - skip
		}
		summaryStartIndex := strings.Index(handText, summarySignifier)

		if summaryStartIndex == -1 {
			log.Printf("missing summary in file %v\n", filename)

			handChan <- handImport{
				filePath: filename,
				hand:     Hand{},
				handErr:  errors.New("no summary found"),
				fileErr:  false,
			}
			continue // the hand lacks important summary data
		}

		summary, parseSummaryErr := parseHandSummary(handText[summaryStartIndex:])

		if parseSummaryErr != nil {
			handChan <- handImport{
				filePath: filename,
				hand:     Hand{},
				handErr:  parseSummaryErr,
				fileErr:  false,
			}
			continue // the hand lacks important summary data
		}

		// update summary.Winners with scanHandLines extracted winners
		summary.Winners = append(summary.Winners, winners...)

		handChan <- handImport{
			filePath: filename,
			hand: Hand{
				Metadata: metadata,
				Players:  players,
				Actions:  actions,
				Summary:  summary,
			},
			handErr: nil,
			fileErr: false,
		}
	}

	if err := fileData.Err(); err != nil {
		return false, fmt.Errorf("Invalid input: %s", err)
	}

	return true, nil
}

// parseHandSummary pulls together the hand summary information and metadata.
func parseHandSummary(summaryText string) (Summary, error) {
	communityCards := parseCommunityCards(summaryText)

	pot, rake, potErr := potFromText(summaryText)
	if potErr != nil {
		return Summary{}, potErr
	}

	summary := Summary{communityCards, pot, rake, []Winner{}}
	return summary, nil
}

func parseMetaData(handText string) (Metadata, error) {
	handID := handIDFromText(handText)
	if handID == "" {
		return Metadata{}, NoHandIDError(fmt.Sprintf("in hand %#v", handText))
	}
	dateTime := parseDateTime(dateTimeFromText(handText))

	btnSeatInt, err := extractButtonSeatFromText(handText)

	if err != nil {
		return Metadata{}, err
	}

	metadata := Metadata{handID, dateTime, int(btnSeatInt)}
	return metadata, nil
}

func extractButtonSeatFromText(handText string) (int64, error) {
	btnSeatString := substringBetween(handText, "Seat #", " is the button")
	btnSeatInt, err := strconv.ParseInt(btnSeatString, 10, 32)
	return btnSeatInt, err
}

// scanHandLines scans the hand data line by line and generates a slice of players, actions and winners. Returns
// a non-nil error if an error was received from the parse helper functions.
func scanHandLines(handText string) ([]Player, []Action, []Winner, error) {

	playersMap := map[string]*Player{}
	var actions []Action
	var winners []Winner
	var street = Preflop
	var order = 0

	scanner := createHandScanner(handText)

	for scanner.Scan() {
		line := scanner.Text()

		actionResult, actionFound, actionErr := parseActionLine(line, &street, &order)

		if actionErr != nil {
			return nil, nil, nil, actionErr
		}

		if actionFound {
			actions = append(actions, actionResult)
		}

		player, playerFound, parsePlayerErr := parsePlayer(line)

		if parsePlayerErr != nil {
			return nil, nil, nil, parsePlayerErr
		}

		if playerFound { // TODO: Cards for hero are overwritten
			updateOrAddPlayer(playersMap, player)
		}

		winner, winnerErr := winnerFromLine(line)
		if winnerErr != nil {
			return nil, nil, nil, winnerErr
		}

		if winner.PlayerName != "" {
			winners = append(winners, winner)
		}

	}

	playersSlice := convertToSlice(playersMap)

	return playersSlice, actions, winners, nil
}

// converToSlice takes a playerMap and returns a []Player ordered by seat position
func convertToSlice(playersMap map[string]*Player) []Player {
	playersSlice := make([]Player, len(playersMap))
	i := 0
	for _, v := range playersMap {
		playersSlice[i] = *v
		i++
	}
	slices.SortFunc(playersSlice, func(a, b Player) int {
		return cmp.Compare(a.Seat, b.Seat)
	})
	return playersSlice
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

func parseCommunityCards(handText string) [2]CommunityCards {
	if strings.Contains(handText, "Hand was run twice") {
		firstBoard := communityCardsFromText(handText, ritFirstBoardSignifier)
		secondBoard := communityCardsFromText(handText, ritSecondBoardSignifier)
		return [2]CommunityCards{firstBoard, secondBoard}
	}
	if strings.Contains(handText, boardSignifier) {
		board := communityCardsFromText(handText, boardSignifier)
		return [2]CommunityCards{board, {}}
	}
	return [2]CommunityCards{}
}

func communityCardsFromText(handText, boardStart string) CommunityCards {
	boardString := substringBetween(handText, boardStart, "]")
	fields := strings.Fields(boardString)

	cc := CommunityCards{}

	if len(fields) >= 3 {
		cc.Flop = [3]Card{Card(fields[0]), Card(fields[1]), Card(fields[2])}
	}

	if len(fields) >= 4 {
		cc.Turn = Card(fields[3])
	}

	if len(fields) >= 5 {
		cc.River = Card(fields[4])
	}

	return cc
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

// actionAmountFromText returns the monetary amount of the action. Returns an error if no currency found.
func actionAmountFromText(line string) (float64, error) {

	if strings.Contains(line, "checks") || strings.Contains(line, "folds") {
		return 0, nil
	}

	return extractAmount(line)
}

func extractAmount(line string) (float64, error) {
	matches := amountRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return 0, CurrencyError(fmt.Sprintf("on line %v", line))
	}

	amount, err := strconv.ParseFloat(matches[1], 64)

	if err != nil {
		return 0, fmt.Errorf("failed parsing amount %s: %w", matches[1], err)
	}

	return amount, nil
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

func parsePlayer(line string) (Player, bool, error) {

	// Found chips, extract name, seat num and chips
	if strings.Contains(line, " in chips)") {
		return extractChipsAndSeatInt(line)
	}

	// Found hero hand extracting name, cards
	if strings.Contains(line, heroHandPrefix) {
		return heroHandFromText(line)
	}

	if strings.Contains(line, showedSignifier) {
		return playerInfoFromText(line, showedSignifier)
	}

	if strings.Contains(line, muckedSignifier) {
		return playerInfoFromText(line, muckedSignifier)
	}

	if strings.Contains(line, foldedSignifier) || strings.Contains(line, collectedSignifier) {
		return playerInfoFromText(line, "")
	}

	return Player{}, false, nil
}

func extractChipsAndSeatInt(line string) (Player, bool, error) {
	seatInt, seatIntErr := seatIntFromText(line)
	playerName := substringBetween(line, ": ", " (")
	chipCount, chipCountErr := extractAmount(line)

	if seatIntErr != nil {
		return Player{}, false, seatIntErr
	}
	if chipCountErr != nil {
		return Player{}, false, chipCountErr
	}
	return Player{
			Username:  playerName,
			Seat:      int(seatInt),
			ChipCount: chipCount,
		},
		true,
		nil
}

func playerInfoFromText(line string, cardPrefix string) (Player, bool, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return Player{}, false, PlayerInfoError(fmt.Sprintf("not enough fields on line %s, expected 4 fields", line))
	}
	playerName := fields[2]
	cards := [2]Card{}

	if cardPrefix != "" {
		cardString := substringBetween(line, cardPrefix, "]")
		before, after, ok := strings.Cut(cardString, " ")
		if ok {
			cards[0] = Card(before)
			cards[1] = Card(after)
		} else {
			return Player{}, false, PlayerInfoError(fmt.Sprintf("not enough fields on line %s, expected 2 fields for cards", line))
		}
	}

	return Player{
			Username: playerName,
			Cards:    cards,
		},
		true,
		nil
}

func heroHandFromText(line string) (Player, bool, error) {
	fields := strings.Fields(line)
	if len(fields) != 5 {
		return Player{}, false, PlayerInfoError(fmt.Sprintf("could not extract hero name, not enough fields on line %s, expected 5 fields", line))
	}
	playerName := fields[2]
	cards := [2]Card{}

	cardString := substringBetween(line, "[", "]")
	before, after, ok := strings.Cut(cardString, " ")
	if ok {
		cards[0] = Card(before)
		cards[1] = Card(after)
	} else {
		return Player{}, false, PlayerInfoError(fmt.Sprintf("not enough fields on line %s, expected 2 fields for cards", line))
	}

	return Player{
			Username: playerName,
			Cards:    cards,
		},
		true,
		nil
}

func winnerFromLine(line string) (Winner, error) {

	triggers := []string{"collected (", " won ("}

	for _, t := range triggers {
		if !strings.Contains(line, t) {
			continue
		}

		amountWithCurrency := substringBetween(line, t, ")")
		amount, amountErr := extractAmount(amountWithCurrency)

		if amountErr != nil {
			return Winner{}, amountErr
		}

		contentBeforeTrigger := substringBetween(line, ": ", t)

		firstSpace := strings.Index(contentBeforeTrigger, " ")
		var playerName string
		if firstSpace == -1 {
			playerName = contentBeforeTrigger
		} else {
			playerName = contentBeforeTrigger[:firstSpace]
		}

		return Winner{
			PlayerName: playerName,
			Amount:     amount,
		}, nil
	}
	return Winner{}, nil
}

func seatIntFromText(line string) (int64, error) {
	matches := seatIntRegex.FindStringSubmatch(line)
	if matches == nil {
		return 0, PlayerInfoError(fmt.Sprintf("no matches for seatInt found on line %v", line))
	}
	if len(matches) != 2 {
		return 0, PlayerInfoError(fmt.Sprintf("failed to extract player seat, multiple matches were found on line %s", line))
	}
	return strconv.ParseInt(matches[1], 10, 32)
}

func parseDateTime(timeString string) time.Time {
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

func potFromText(handText string) (float64, float64, error) {

	if strings.Contains(handText, potSizeSignifier) {
		potString, rakeString, _ := strings.Cut(handText, "|")

		potSize, potErr := extractAmount(potString)
		rake, rakeErr := extractAmount(rakeString)
		if potErr != nil {
			return 0, 0, fmt.Errorf("amountFromText: unable to parse float parsing: %w", potErr)
		}
		if rakeErr != nil {
			return 0, 0, fmt.Errorf("amountFromText: unable to parse float parsing: %w", rakeErr)
		}
		return potSize, rake, nil
	}
	return 0, 0, nil
}

func updateOrAddPlayer(players map[string]*Player, player Player) {
	if p, ok := players[player.Username]; ok {
		if p.Cards == [2]Card{"", ""} {
			p.Cards = player.Cards
		}
	} else {
		newPlayer := player
		players[player.Username] = &newPlayer
	}
}

// subStringBetween returns the substring between the first instance of characters start and end.
// If text does not contain either the start or end string, the original text is returned unchanged.
func substringBetween(text, start, end string) string {
	startIndex := strings.Index(text, start)
	if startIndex == -1 {
		return text
	}
	startSubString := text[startIndex+len(start):]
	before, _, ok := strings.Cut(startSubString, end)

	if !ok || startIndex+len(start) > len(text) {
		return text
	}

	return before
}

// CreateHandScanner returns a pointer to bufio.Scanner for parsing Hand data
func createHandScanner(h string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(h))
	return scanner
}

func splitByHands() func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	delimiter := []byte(handInfoDelimiter)
	delimLen := len(delimiter)

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		dataLen := len(data)
		if atEOF && dataLen == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, delimiter); i >= 0 {
			return i + delimLen, data[0:i], nil
		}

		if atEOF {
			return dataLen, data, nil
		}
		return 0, nil, nil
	}
}
