package pokerhud

import (
	"bufio"
	"io/fs"
	"slices"
	"strings"
)

const handInfoDelim string = "\n\n\n"

type Hand struct {
	Id        string
	Players   []string
	HeroCards string
}

// Todo - At some point we're going to want to make an interface of sorts so that we handle Pokerstars hands, Party poker hands. etc.
//so Hand will actually be an interface and there'll be a pokerstarshand struct that implements hand interface... methods TBD.
// equally each hand

// Do we need Action struct? Does that belong internally to the hand

// We'd need some kind of continual running to scan the FS for new hand files and keep reading from the same particular one.

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
	sessionData := string(fileData)
	handsText := strings.Split(sessionData, handInfoDelim)

	var hands []Hand
	for _, h := range handsText {

		handId := handIdFromText(h)

		scanner := bufio.NewScanner(strings.NewReader(h))
		scanner.Scan()
		scanner.Scan()

		playerNames := playerNamesFromText(scanner)
		heroCards := heroCardsFromText(scanner)

		hands = append(hands, Hand{
			Id:        handId,
			Players:   playerNames,
			HeroCards: heroCards,
		})
	}

	return hands
}

func handIdFromText(h string) string {
	return strings.Split(strings.Split(h, ":")[0], "#")[1]
}

func playerNamesFromText(scanner *bufio.Scanner) []string {
	var playerNames []string

	//TODO - possibly split this logic? e.g. for scanner.scan() and then if string contains "seat "
	for scanner.Scan() && strings.Contains(scanner.Text(), "Seat ") {
		start := strings.Index(scanner.Text(), ": ")
		end := strings.Index(scanner.Text(), " (")

		if start != -1 && end != -1 {
			playerNames = append(playerNames, scanner.Text()[start+2:end])
		}
	}
	return playerNames
}

func heroCardsFromText(scanner *bufio.Scanner) string {

	var heroCards string
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() && strings.Contains(scanner.Text(), "Dealt to") {
		start := strings.Index(scanner.Text(), "[")
		end := strings.Index(scanner.Text(), "]")
		if start != -1 && end != -1 {
			heroCards = scanner.Text()[start+1 : end]
			return heroCards
		}

	}
	return heroCards
}
