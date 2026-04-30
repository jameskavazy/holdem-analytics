package hands

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"slices"
	"strconv"
	"time"
)

// Delimiter and signifier constants for parsing hand files
var (
	handInfoDelimiter       = []byte("\nPokerStars ")
	newLine                 = []byte("\n")
	flopSignifier           = []byte("*** FLOP ***")
	turnSignifier           = []byte("*** TURN ***")
	riverSignifier          = []byte("*** RIVER ***")
	summarySignifier        = []byte("*** SUMMARY ***")
	heroHandPrefix          = []byte("Dealt to")
	showedSignifier         = []byte("showed [")
	muckedSignifier         = []byte("mucked [")
	foldedSignifier         = []byte("folded")
	collectedSignifier      = []byte("collected (")
	boardSignifier          = []byte("Board [")
	ritFirstBoardSignifier  = []byte("FIRST Board [")
	ritSecondBoardSignifier = []byte("SECOND Board [")
	potSizeSignifier        = []byte("Total pot ")

	// Action signifiers
	sigFolds  = []byte(" folds")
	sigChecks = []byte(" checks")
	sigCalls  = []byte(" calls")
	sigBets   = []byte(" bets")
	sigRaises = []byte(" raises")
	sigPosts  = []byte(" posts")
)

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
		handBytes := fileData.Bytes()

		metadata, metadataErr := parseMetaData(handBytes)

		if metadataErr != nil {
			handChan <- handImport{filename, Hand{}, metadataErr, false}
			continue // the hand lacks crucial metadata - skip
		}

		players, actions, winners, scanHandErr := scanHandLines(handBytes)
		if scanHandErr != nil {
			handChan <- handImport{
				filePath: filename,
				hand:     Hand{},
				handErr:  scanHandErr,
				fileErr:  false,
			}
			continue // the hand lacks crucial gameplay info - skip
		}
		summaryStartIndex := bytes.Index(handBytes, summarySignifier)

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

		summary, parseSummaryErr := parseHandSummary(handBytes[summaryStartIndex:])

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
func parseHandSummary(summaryText []byte) (Summary, error) {
	communityCards := parseCommunityCards(summaryText)

	pot, rake, potErr := potFromText(summaryText)
	if potErr != nil {
		return Summary{}, potErr
	}

	summary := Summary{communityCards, pot, rake, []Winner{}}
	return summary, nil
}

func parseMetaData(handText []byte) (Metadata, error) {
	handID := handIDFromText(handText)
	if handID == nil {
		return Metadata{}, NoHandIDError(fmt.Sprintf("in hand %#v", handText))
	}
	dateTime := parseDateTime(dateTimeFromText(handText))

	btnSeatInt, err := extractButtonSeatFromText(handText)

	if err != nil {
		return Metadata{}, err
	}

	metadata := Metadata{string(handID), dateTime, int(btnSeatInt)}
	return metadata, nil
}

func extractButtonSeatFromText(handBytes []byte) (int64, error) {
	btnSeatString := substringBetween(handBytes, []byte("Seat #"), []byte(" is the button"))
	btnSeatInt, err := strconv.ParseInt(string(btnSeatString), 10, 32)
	return btnSeatInt, err
}

// scanHandLines scans the hand data line by line and generates a slice of players, actions and winners. Returns
// a non-nil error if an error was received from the parse helper functions.
func scanHandLines(handText []byte) ([]Player, []Action, []Winner, error) {

	playersMap := map[string]Player{}
	var actions []Action
	var winners []Winner
	var street = Preflop
	var order = 0

	lines := bytes.SplitSeq(handText, newLine)
	for line := range lines {
		if len(line) == 0 {
			continue
		}

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

		if playerFound {
			updateOrAddPlayer(playersMap, player)
		}

		winner, winnerErr := winnerFromLine(line)
		if winnerErr != nil {
			return nil, nil, nil, winnerErr
		}

		winners = append(winners, winner...)
	}

	playersSlice := convertToSlice(playersMap)

	return playersSlice, actions, winners, nil
}

// converToSlice takes a playerMap and returns a []Player ordered by seat position
func convertToSlice(playersMap map[string]Player) []Player {
	playersSlice := make([]Player, len(playersMap))
	i := 0
	for _, v := range playersMap {
		playersSlice[i] = v
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
func parseActionLine(line []byte, actionStreet *Street, order *int) (Action, bool, error) {
	actionStreet.next(line)
	actionType, actionFound := actionTypeFromText(line)

	if !actionFound {
		return Action{}, actionFound, nil
	}

	playerName, playerErr := actionPlayerNameFromText(line)

	actionStartIdx := bytes.Index(line, []byte(": "))
	amount, amtErr := actionAmountFromText(line[actionStartIdx:])

	if playerErr != nil {
		return Action{}, actionFound, ActionParseError(fmt.Sprintf("%v %v", playerErr, line))
	}

	if amtErr != nil {
		return Action{}, actionFound, ActionParseError(fmt.Sprintf("%v %v", amtErr, line))
	}

	*order++

	return Action{
		ActionType: actionType,
		PlayerName: string(playerName),
		Street:     *actionStreet,
		Order:      *order,
		Amount:     amount,
	}, actionFound, nil
}

func parseCommunityCards(handText []byte) [2]CommunityCards {
	if bytes.Contains(handText, []byte("Hand was run twice")) {
		firstBoard := communityCardsFromText(handText, ritFirstBoardSignifier)
		secondBoard := communityCardsFromText(handText, ritSecondBoardSignifier)
		return [2]CommunityCards{firstBoard, secondBoard}
	}
	if bytes.Contains(handText, boardSignifier) {
		board := communityCardsFromText(handText, boardSignifier)
		return [2]CommunityCards{board, {}}
	}
	return [2]CommunityCards{}
}

func communityCardsFromText(handText, boardStart []byte) CommunityCards {
	boardString := substringBetween(handText, boardStart, []byte("]"))
	fields := bytes.Fields(boardString)

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

func actionTypeFromText(line []byte) (ActionType, bool) {
	switch {
	case bytes.Contains(line, sigFolds):
		return ActionFold, true
	case bytes.Contains(line, sigRaises):
		return ActionRaise, true
	case bytes.Contains(line, sigCalls):
		return ActionCall, true
	case bytes.Contains(line, sigBets):
		return ActionBet, true
	case bytes.Contains(line, sigChecks):
		return ActionCheck, true
	case bytes.Contains(line, sigPosts):
		return ActionPost, true
	default:
		return "", false
	}
}

func actionPlayerNameFromText(line []byte) ([]byte, error) {
	before, _, found := bytes.Cut(line, []byte(":"))
	if !found {
		return nil, errors.New("could not parse name in action")
	}
	return before, nil
}

// actionAmountFromText returns the monetary amount of the action. Returns an error if no currency found.
func actionAmountFromText(line []byte) (float64, error) {

	if bytes.Contains(line, []byte("checks")) || bytes.Contains(line, []byte("folds")) {
		return 0, nil
	}

	return extractAmount(line)
}

func extractAmount(line []byte) (float64, error) {

	i := bytes.IndexByte(line, '$')

	if i == -1 {
		return 0, CurrencyError(fmt.Sprintf("on line %v", string(line)))
	}

	j := i + 1

	for j < len(line) && (line[j] == '.' || (line[j] >= '0' && line[j] <= '9')) {
		j++
	}

	return strconv.ParseFloat(string(line[i+1:j]), 64)
}

// handIDFromText returns the hand ID string from the hand info string
func handIDFromText(handText []byte) []byte {
	if !bytes.Contains(handText, []byte("Hand #")) {
		return nil
	}

	if bytes.Contains(handText, []byte(":")) {
		return substringBetween(handText, []byte("#"), []byte(":"))
	}

	return nil
}

func parsePlayer(line []byte) (Player, bool, error) {

	// Found chips, extract name, seat num and chips
	if bytes.Contains(line, []byte(" in chips)")) {
		return extractChipsAndSeatInt(line)
	}

	// Found hero hand extracting name, cards
	if bytes.Contains(line, heroHandPrefix) {
		return heroHandFromText(line)
	}

	if bytes.Contains(line, showedSignifier) {
		return playerInfoFromText(line, showedSignifier)
	}

	if bytes.Contains(line, muckedSignifier) {
		return playerInfoFromText(line, muckedSignifier)
	}

	if bytes.Contains(line, foldedSignifier) || bytes.Contains(line, collectedSignifier) {
		return playerInfoFromText(line, nil)
	}

	return Player{}, false, nil
}

func extractChipsAndSeatInt(line []byte) (Player, bool, error) {
	seatInt, seatIntErr := seatIntFromText(line)
	playerName := substringBetween(line, []byte(": "), []byte(" ("))

	chipCountIdx := bytes.Index(line, []byte("($"))
	chipCount, chipCountErr := extractAmount(line[chipCountIdx:])

	if seatIntErr != nil {
		return Player{}, false, seatIntErr
	}
	if chipCountErr != nil {
		return Player{}, false, chipCountErr
	}
	return Player{
			Username:  string(playerName),
			Seat:      int(seatInt),
			ChipCount: chipCount,
		},
		true,
		nil
}

func playerInfoFromText(line []byte, cardPrefix []byte) (Player, bool, error) {
	playerName := substringBetween(line, []byte(": "), []byte(" "))

	cards := [2]Card{}

	if cardPrefix != nil {
		cardString := substringBetween(line, cardPrefix, []byte("]"))
		before, after, ok := bytes.Cut(cardString, []byte(" "))
		if ok {
			cards[0] = Card(before)
			cards[1] = Card(after)
		} else {
			return Player{}, false, PlayerInfoError(fmt.Sprintf("not enough fields on line %s, expected 2 fields for cards", string(line)))
		}
	}

	return Player{
			Username: string(playerName),
			Cards:    cards,
		},
		true,
		nil
}

func heroHandFromText(line []byte) (Player, bool, error) {
	playerName := substringBetween(line, []byte("Dealt to "), []byte(" ["))
	cards := [2]Card{}

	cardString := substringBetween(line, []byte("["), []byte("]"))
	before, after, ok := bytes.Cut(cardString, []byte(" "))
	if ok {
		cards[0] = Card(before)
		cards[1] = Card(after)
	} else {
		return Player{}, false, PlayerInfoError(fmt.Sprintf("not enough fields on line %s, expected 2 fields for cards", string(line)))
	}

	return Player{
			Username: string(playerName),
			Cards:    cards,
		},
		true,
		nil
}

func winnerFromLine(line []byte) ([]Winner, error) {
	var winners []Winner
	triggers := [][]byte{[]byte("collected ("), []byte(" won (")}

	for _, t := range triggers {
		if !bytes.Contains(line, t) {
			continue
		}

		contentBeforeTrigger := substringBetween(line, []byte(": "), t)

		// Multiple Winners branch
		if c := bytes.Count(line, t); c > 1 {
			first, second, ok := bytes.Cut(line, []byte(","))
			if !ok {
				return []Winner{}, fmt.Errorf("on no %w", ErrPlayerInfo)
			}
			fw, fwErr := winnerFromLine(first)
			if fwErr != nil {
				return []Winner{}, fwErr
			}
			winners = append(winners, fw...)

			// second is the tail after the comma, missing the player prefix — reconstruct
			// line by prepending the player name extracted from the first clause
			secondWithPlayerName := bytes.Join([][]byte{contentBeforeTrigger, second}, []byte(" "))
			sw, swErr := winnerFromLine(secondWithPlayerName)
			if fwErr != nil {
				return []Winner{}, swErr
			}

			winners = append(winners, sw...)
			return winners, nil
		}

		amountWithCurrency := substringBetween(line, t, []byte(")"))
		amount, amountErr := extractAmount(amountWithCurrency)

		if amountErr != nil {
			return []Winner{}, amountErr
		}

		before, _, ok := bytes.Cut(contentBeforeTrigger, []byte(" "))
		var playerName []byte
		if !ok {
			playerName = contentBeforeTrigger
		} else {
			playerName = before
		}

		winners = append(winners, Winner{
			PlayerName: string(playerName),
			Amount:     amount,
		})
	}
	return winners, nil
}

func seatIntFromText(line []byte) (int64, error) {
	if !bytes.HasPrefix(line, []byte("Seat ")) {
		return 0, PlayerInfoError(fmt.Sprintf("no matches for seatInt found on line %v", string(line)))
	}

	i := len("Seat ")
	j := i
	for j < len(line) && line[j] >= '0' && line[j] <= '9' {
		j++
	}

	return strconv.ParseInt(string(line[i:j]), 10, 32)

}

func parseDateTime(timeString []byte) time.Time {
	siteTime, _ := time.ParseInLocation(time.DateTime, string(timeString), siteLocation)
	return siteTime.Local()
}

// dateTimeFromText extracts the relevant time from the hand information
// and converts to a string that can be transformed into a time.Time
func dateTimeFromText(line []byte) []byte {

	var timeString []byte
	if bytes.ContainsAny(line, "[]") {
		timeString = substringBetween(line, []byte("["), []byte(" ET]"))
	}

	for i := range timeString {
		if timeString[i] == '/' {
			timeString[i] = '-'
		}
	}
	return timeString
}

func potFromText(handBytes []byte) (float64, float64, error) {

	if bytes.Contains(handBytes, potSizeSignifier) {
		potString, rakeString, _ := bytes.Cut(handBytes, []byte("|"))

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

func updateOrAddPlayer(players map[string]Player, player Player) {
	if p, ok := players[player.Username]; ok {
		if p.Cards[0] == "" {
			p.Cards = player.Cards
			players[player.Username] = p
		}
	} else {
		newPlayer := player
		players[player.Username] = newPlayer
	}
}

// subStringBetween returns the substring between the first instance of characters start and end.
// If text does not contain either the start or end string, the original text is returned unchanged.
func substringBetween(text, start, end []byte) []byte {
	startIndex := bytes.Index(text, start)
	if startIndex == -1 {
		return text
	}
	startSubString := text[startIndex+len(start):]
	before, _, ok := bytes.Cut(startSubString, end)

	if !ok || startIndex+len(start) > len(text) {
		return text
	}

	return before
}

func splitByHands() func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	delimiter := handInfoDelimiter
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
