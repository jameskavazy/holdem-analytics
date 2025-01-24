package pokerhud

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestActionTypeFromText(t *testing.T) {

	cases := map[string]string{
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

func TestBuildActions(t *testing.T) {
	reader := strings.NewReader(testHand)
	scanner := bufio.NewScanner(reader)
	var actions []Action
	var got []Action
	for scanner.Scan() {
		got = append(got, BuildActions(scanner, actions)...)
	}
	want := []Action{
		{
			// Player:     "KavarzE",
			ActionType: "posts",
		},
		{
			// Player:     "getaddicted",
			ActionType: "posts",
		},
		{
			// Player:     "Mythic Max:",
			ActionType: "folds",
		},
		{
			// Player:     "Cl8rker",
			ActionType: "folds",
		},
		{
			// Player:     "MGPN",
			ActionType: "raises",
		},
		{
			// Player:     "SyraXmaX",
			ActionType: "folds",
		},
		{
			// Player:     "KavarzE",
			ActionType: "folds",
		},
		{
			// Player:     "getaddicted",
			ActionType: "folds",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, wanted %#v", got, want)
	}
}

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

const testHand string = `KavarzE: posts small blind $0.02
getaddicted: posts big blind $0.05
Mythic Max: folds
Cl8rker: folds
MGPN: raises $0.05 to $0.10
SyraXmaX: folds
KavarzE: folds
getaddicted: folds`
