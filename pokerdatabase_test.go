package pokerhud

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestActionTypeFromText(t *testing.T) {
	cases := map[string]ActionType{
		"kv_def: posts small blind $0.02": "posts",
		"KavarzE: posts big blind $0.05":  "posts",
		"arsad725: folds":                 "folds",
		"RE0309: calls $0.05":             "calls",
		"pernadao1599: calls $0.05":       "calls",
		"maximoIV: folds":                 "folds",
		"dlourencobss: calls $0.03":       "calls",
		"KavarzE: checks":                 "checks",
		"dlourencobss: bets $0.10":        "bets",
		"KavarzE: folds":                  "folds",
		"RE0309: folds":                   "folds",
		"pernadao1599: calls $0.10":       "calls",
		"dlourencobss: bets $0.27":        "bets",
		"pernadao1599: calls $0.27":       "calls",
		"dlourencobss: checks":            "checks",
		"pernadao1599: checks":            "checks",
	}

	for c, want := range cases {
		buffer := bytes.Buffer{}
		fmt.Fprintf(&buffer, "Post Scenario: %v", c)

		t.Run(buffer.String(), func(t *testing.T) {
			scanner := createHandScanner(c)
			scanner.Scan()
			got, _ := actionTypeFromText(scanner)
			if got != want {
				t.Errorf("got %v, but wanted %v", got, want)
			}
		})
	}
}

func TestPlayerNameActionFromText(t *testing.T) {
	cases := map[string]string{
		"kv_def: posts small blind $0.02": "kv_def",
		"KavarzE: posts big blind $0.05":  "KavarzE",
		"arsad725: folds":                 "arsad725",
		"RE0309: calls $0.05":             "RE0309",
		"pernadao1599: calls $0.05":       "pernadao1599",
		"maximoIV: folds":                 "maximoIV",
		"dlourencobss: calls $0.03":       "dlourencobss",
		"KavarzE: checks":                 "KavarzE",
		"dlourencobss: bets $0.10":        "dlourencobss",
		"KavarzE: folds":                  "KavarzE",
		"RE0309: folds":                   "RE0309",
		"pernadao1599: calls $0.10":       "pernadao1599",
		"dlourencobss: bets $0.27":        "dlourencobss",
		"pernadao1599: calls $0.27":       "pernadao1599",
		"dlourencobss: checks":            "dlourencobss",
		"pernadao1599: checks":            "pernadao1599",
	}

	for c, want := range cases {
		buffer := bytes.Buffer{}
		fmt.Fprintf(&buffer, "Post Scenario: %v", c)

		t.Run(buffer.String(), func(t *testing.T) {
			scanner := createHandScanner(c)
			scanner.Scan()
			got, _ := actionPlayerNameFromText(scanner)
			if got != want {
				t.Errorf("got %v, but wanted %v", got, want)
			}
		})
	}
}

// func TestHandsFromSessionFile(t *testing.T) {
// 	fileSystem := fstest.MapFS{
// 		"failed": {Data: nil},
// 	}

// 	got, _ := handsFromSessionFile(fileSystem, "failed")
// 	if got != nil {
// 		t.Error("oh god")
// 	}

// }

func TestActionAmountFromText(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cases := map[string]float64{
			"kv_def: posts small blind $0.02": 0.02,
			"KavarzE: posts big blind $0.05":  0.05,
			"arsad725: folds":                 0,
			"RE0309: calls $0.05":             0.05,
			"pernadao1599: calls $0.05":       0.05,
			"maximoIV: folds":                 0,
			"dlourencobss: calls $0.03":       0.03,
			"KavarzE: checks":                 0,
			"dlourencobss: bets $0.10":        0.1,
			"KavarzE: folds":                  0,
			"RE0309: folds":                   0,
			"pernadao1599: calls $0.10":       0.1,
			"dlourencobss: bets $0.27":        0.27,
			"pernadao1599: calls $0.27":       0.27,
			"dlourencobss: checks":            0,
			"pernadao1599: checks":            0,
			"KavarzE: raises $0.08 to $0.13":  0.08,
		}

		for c, want := range cases {
			buffer := bytes.Buffer{}
			fmt.Fprintf(&buffer, "Scenario: %v", c)

			t.Run(buffer.String(), func(t *testing.T) {
				scanner := testingScanner(c)
				got, _ := actionAmountFromText(scanner)

				if got != want {
					t.Errorf("got %v, but wanted %v", got, want)
				}
			})
		}
	})

	t.Run("error pathway", func(t *testing.T) {
		cases := []string{
			"kv_def: bets small blind 0.02",
			"KavarzE: posts big blind 0.05",
			"KavarzE:  big blind 0.05",
		}

		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				_, err := actionAmountFromText(testingScanner(c))
				want := CurrencyError(fmt.Sprintf("on line %v", c)).Error()

				if err.Error() != want {
					t.Fatalf("expected \"%v\" error but got \"%v\"", CurrencyError(fmt.Sprintf("on line %v", c)), err)
				}
			})
		}

	})
}

func TestHandsFromSessionFile(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"zoom.txt": {Data: []byte(`PokerStars Hand #123: blah blah
Seat 1: test ($6000 in chips)
Seat 2: test2 ($3000 in chips)
Dealt to me [Ad Ac]
KavarzE: bets $2.33`)},
		}

		got, err := handsFromSessionFile(fileSystem, "zoom.txt")

		want := []Hand{
			{
				"123", []Player{{Username: "test"}, {Username: "test2"}}, "Ad Ac", []Action{
					{Player{Username: "KavarzE"}, 1, Preflop, Bets, 2.33},
				},
			},
		}

		if err != nil {
			t.Fatal("got an error but didn't expect one: ", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v wanted %#v", got, want)
		}
	})

	t.Run("error pathway", func(t *testing.T) {
		fileSystem := failingFS{}

		_, err := handsFromSessionFile(fileSystem, "zoom.txt")

		if err == nil {
			fmt.Print(err)
			t.Fatal("expected an error but didn't get one!")
		}
	})
}

func TestParseHandData(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		handData := []byte(`PokerStars Hand #123: blah blah
Seat 1: test ($6000 in chips)
Seat 2: test2 ($3000 in chips)
Dealt to me [Ad Ac]
KavarzE: bets $2.33`)

		got, _ := parseHandData(handData)
		want := []Hand{
			{
				"123", []Player{{Username: "test"}, {Username: "test2"}}, "Ad Ac", []Action{{Player{Username: "KavarzE"}, 1, Preflop, Bets, 2.33}},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, wanted %#v", got, want)
		}
	})

	t.Run("Random non-hand data", func(t *testing.T) {
		handData := []byte(`Random non-hand data, whoops!`)

		_, err := parseHandData(handData)

		if err == nil {
			t.Errorf("expected an error but didn't get one")
		}

		if !errors.Is(ErrNoHandID, err) {
			t.Errorf("expected a %v type of error but got a different one: %v", ErrNoHandID, err)
		}
	})
}

func TestHandPlayerNames(t *testing.T) {
	handData := "Seat 2: test2 ($3000 in chips)"
	scanner := testingScanner(handData)
	got := handPlayerNameFromText(scanner)
	want := "test2"

	if got != want {
		t.Errorf("got %v wanted %v", got, want)
	}
}

func TestSetHeroCards(t *testing.T) {
	handData := "Dealt to Karv [Ac Kc]"
	scanner := testingScanner(handData)
	got := heroCardsFromText(scanner)
	want := "Ac Kc"

	if got != want {
		t.Errorf("got %v wanted %v", got, want)
	}
}

func TestHandIdFromText(t *testing.T) {
	handData := "Pokerstars #6548679821301346841: Holdem don't care"

	got := handIDFromText(handData)
	want := "6548679821301346841"

	if got != want {
		t.Errorf("got %v wanted %v", got, want)
	}
}

func TestParseAction(t *testing.T) {
	dummyActions := []Action{
		{Player{"Kavarz"}, 2, Flop, Bets, 3},
		{Player{"Burty"}, 3, Flop, Calls, 3},
	}
	var dummyStreet = Flop
	order := 4

	handData := "kv_def: calls $3"
	scanner := testingScanner(handData)

	got, err := parseAction(scanner, dummyActions, &dummyStreet, &order)
	if err != nil {
		t.Error(err)
	}

	want := []Action{
		{Player{"Kavarz"}, 2, Flop, Bets, 3},
		{Player{"Burty"}, 3, Flop, Calls, 3},
		{Player{"kv_def"}, 4, Flop, Calls, 3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %#v but got %#v", want, got)
	}
}

func testingScanner(handData string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(handData))
	scanner.Scan()
	return scanner
}

type failingFS struct{}

func (f failingFS) Open(name string) (fs.File, error) {
	return nil, errors.New("oh no i always fail")
}

// func actionBuildHelper(player string, actionType ActionType, street Street, order int, amount float64) Action {
// 	return Action{
// 		Player:     player,
// 		ActionType: actionType,
// 		Street:     street,
// 		Order:      order,
// 		Amount:     amount,
// 	}
// }

// func TestStreetActionFromText(t *testing.T) {
// 	cases := map[string]string{

// 	}

// }

// const testHands string = `PokerStars Zoom Hand #254489598204:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:51:32 WET [2025/01/21 15:51:32 ET]
// Table 'Donati' 6-max Seat #1 is the button
// Seat 1: JDfq28 ($5.11 in chips)
// Seat 2: kv_def ($11.57 in chips)
// Seat 3: KavarzE ($5 in chips)
// Seat 4: MGPN ($4.63 in chips)
// Seat 5: ikin23 ($7.63 in chips)
// Seat 6: honda589 ($5.38 in chips)
// kv_def: posts small blind $0.02
// KavarzE: posts big blind $0.05
// *** HOLE CARDS ***
// Dealt to KavarzE [3c 7c]
// MGPN: folds
// ikin23: folds
// honda589: folds
// JDfq28: raises $0.07 to $0.12
// kv_def: calls $0.10
// KavarzE: calls $0.07
// *** FLOP *** [Qc As 3d]
// kv_def: checks
// KavarzE: checks
// JDfq28: checks
// *** TURN *** [Qc As 3d] [2h]
// kv_def: bets $0.20
// KavarzE: folds
// JDfq28: folds
// Uncalled bet ($0.20) returned to kv_def
// kv_def collected $0.35 from pot
// kv_def: doesn't show hand
// *** SUMMARY ***
// Total pot $0.36 | Rake $0.01
// Board [Qc As 3d 2h]
// Seat 1: JDfq28 (button) folded on the Turn
// Seat 2: kv_def (small blind) collected ($0.35)
// Seat 3: KavarzE (big blind) folded on the Turn
// Seat 4: MGPN folded before Flop (didn't bet)
// Seat 5: ikin23 folded before Flop (didn't bet)
// Seat 6: honda589 folded before Flop (didn't bet)

// const testHand string = `KavarzE: posts small blind $0.02
// getaddicted: posts big blind $0.05
// Mythic Max: folds
// Cl8rker: folds
// MGPN: raises $0.05 to $0.10
// SyraXmaX: folds
// KavarzE: folds
// getaddicted: folds`
