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
	"time"
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
				"123", time.Time{}.Local(), []Player{{Username: "test"}, {Username: "test2"}}, "Ad Ac", []Action{
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
				"123", time.Time{}.Local(), []Player{{Username: "test"}, {Username: "test2"}}, "Ad Ac", []Action{{Player{Username: "KavarzE"}, 1, Preflop, Bets, 2.33}},
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

		if !errors.Is(ErrNoHandID, err[0]) {
			t.Errorf("expected a %v type of error but got a different one: %v", ErrNoHandID, err)
		}
	})

	t.Run("file with 3 hands, but one is corrupted", func(t *testing.T) {
		handData := []byte(brokenHands)

		hands, err := parseHandData(handData)

		if len(hands) != 2 {
			// fmt.Printf("%#v\n\n %#v", hands, err)
			t.Errorf("wanted 2 hands and 1 error but got %#v hands", len(hands))
		}

		if len(err) != 1 {
			t.Errorf("wanted 1 error but got %v", len(err))
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
	handData := "Pokerstars Hand #6548679821301346841: Holdem don't care"

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

func TestDateFromHandText(t *testing.T) {
	cases := []struct {
		test string
		want string
	}{
		{"PokerStars Zoom Hand #254489598204:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:51:32 WET [2025/01/21 15:51:32 ET]", "2025-01-21 15:51:32"},
		{"PokerStars Zoom Hand #254489608193:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:52:05 WET [2025/01/21 15:52:05 ET]", "2025-01-21 15:52:05"},
		{"PokerStars Zoom Hand #254489609065:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:52:09 WET [2025/01/21 15:52:09 ET]", "2025-01-21 15:52:09"},
		{"PokerStars Zoom Hand #254489686769:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:56:56 WET [2025/01/21 15:56:56 ET]", "2025-01-21 15:56:56"},
		{"PokerStars Hand #254581458091:  Hold'em No Limit ($0.02/$0.05 USD) - 2025/01/27 17:49:38 WET [2025/01/27 12:49:38 ET]", "2025-01-27 12:49:38"},
		{"PokerStars Hand #254581458091:  Hold'em No Limit ($0.02/$0.05 USD) - [2025/01/27 12:49:38 ET]", "2025-01-27 12:49:38"},
		{"PokerStars Hand #254581458091:  Hold'em No Limit ($0.02/$0.05 USD) - ", ""},
	}

	for _, tt := range cases {
		t.Run(tt.test, func(t *testing.T) {

			got := dateTimeStringFromHandText(tt.test)
			if got != tt.want {
				t.Errorf("got %v but wanted %v", got, tt.want)
			}

		})
	}
}

func TestParseDateTime(t *testing.T) {
	info := "2025-01-27 12:49:38"

	got := parseDateTime(info)
	localTime, _ := time.Parse(time.DateTime, "2025-01-27 17:49:38")
	want := localTime.Local()

	if got != want {
		t.Errorf("got %v wanted %v", got, want)
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


const brokenHands string = `PokerStars Hand #174088855475:  Hold'em No Limit (50/100) - 2017/08/08 23:16:30 MSK [2017/08/08 16:16:30 ET]
Table 'Euphemia II' 6-max (Play Money) Seat #3 is the button
Seat 1: adevlupec (53368 in chips) 
Seat 2: Dette32 (10845 in chips) 
Seat 3: Drug08 (9686 in chips) 
Seat 4: FluffyStutt (11326 in chips) 
FluffyStutt: posts small blind 50
adevlupec: posts big blind 100
*** HOLE CARDS ***
Dealt to FluffyStutt [2h Ks]
FluffyStutt said, "nh"
Dette32: calls 100
Drug08: calls 100
FluffyStutt: folds 
adevlupec: checks 
*** FLOP *** [8h 7s 8d]
adevlupec: checks 
Dette32: checks 
Drug08: checks 
*** TURN *** [8h 7s 8d] [Th]
adevlupec: checks 
Dette32: checks 
Drug08: checks 
*** RIVER *** [8h 7s 8d Th] [2c]
adevlupec: checks 
Dette32: checks 
Drug08: checks 
*** SHOW DOWN ***
adevlupec: shows [Qs Ts] (two pair, Tens and Eights)
Dette32: mucks hand 
Drug08: mucks hand 
adevlupec collected 332 from pot
*** SUMMARY ***
Total pot 350 | Rake 18 
Board [8h 7s 8d Th 2c]
Seat 1: adevlupec (big blind) showed [Qs Ts] and won (332) with two pair, Tens and Eights
Seat 2: Dette32 mucked [5s Kc]
Seat 3: Drug08 (button) mucked [4d 6h]
Seat 4: FluffyStutt (small blind) folded before Flop



Pokerstars broken hand with bad data somehow...
Table 'Euphemia II' 6-max (Play Money) Seat #4 is the button
Seat 1: adevlupec (53600 in chips) 
Seat 2: Dette32 (10745 in chips) 
Seat 3: Drug08 (9586 in chips) 
Seat 4: FluffyStutt (11276 in chips) 
yanksea will be allowed to play after the button
adevlupec: posts small blind 50
Dette32: posts big blind 100
*** HOLE CARDS ***
Dealt to FluffyStutt [8s Qc]
Drug08: calls 100
FluffyStutt: calls 100
adevlupec: folds 
Dette32: checks 
*** FLOP *** [8c Qh 4c]
Dette32: checks 
Drug08: checks 
FluffyStutt: bets 332
Dette32: folds 
Drug08: calls 332
*** TURN *** [8c Qh 4c] [Ac]
Drug08: checks 
FluffyStutt: bets 963
Drug08: calls 963
*** RIVER *** [8c Qh 4c Ac] [Td]
Drug08: checks 
FluffyStutt: bets 9881 and is all-in
Drug08: folds 
Uncalled bet (9881) returned to FluffyStutt
FluffyStutt collected 2793 from pot
FluffyStutt: doesn't show hand 
*** SUMMARY ***
Total pot 2940 | Rake 147 
Board [8c Qh 4c Ac Td]
Seat 1: adevlupec (small blind) folded before Flop
Seat 2: Dette32 (big blind) folded on the Flop
Seat 3: Drug08 folded on the River
Seat 4: FluffyStutt (button) collected (2793)



PokerStars Hand #174088919486:  Hold'em No Limit (50/100) - 2017/08/08 23:17:57 MSK [2017/08/08 16:17:57 ET]
Table 'Euphemia II' 6-max (Play Money) Seat #4 is the button
Seat 1: adevlupec (53600 in chips) 
Seat 2: Dette32 (10745 in chips) 
Seat 3: Drug08 (9586 in chips) 
Seat 4: FluffyStutt (11276 in chips) 
yanksea will be allowed to play after the button
adevlupec: posts small blind 50
Dette32: posts big blind 100
*** HOLE CARDS ***
Dealt to FluffyStutt [8s Qc]
Drug08: calls 100
FluffyStutt: calls 100
adevlupec: folds 
Dette32: checks 
*** FLOP *** [8c Qh 4c]
Dette32: checks 
Drug08: checks 
FluffyStutt: bets 332
Dette32: folds 
Drug08: calls 332
*** TURN *** [8c Qh 4c] [Ac]
Drug08: checks 
FluffyStutt: bets 963
Drug08: calls 963
*** RIVER *** [8c Qh 4c Ac] [Td]
Drug08: checks 
FluffyStutt: bets 9881 and is all-in
Drug08: folds 
Uncalled bet (9881) returned to FluffyStutt
FluffyStutt collected 2793 from pot
FluffyStutt: doesn't show hand 
*** SUMMARY ***
Total pot 2940 | Rake 147 
Board [8c Qh 4c Ac Td]
Seat 1: adevlupec (small blind) folded before Flop
Seat 2: Dette32 (big blind) folded on the Flop
Seat 3: Drug08 folded on the River
Seat 4: FluffyStutt (button) collected (2793)`