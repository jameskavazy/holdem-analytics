package pokerhud

import (
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
			got, _ := actionTypeFromText(c)
			if got != want {
				t.Errorf("got %v, but wanted %v", got, want)
			}
		})
	}
}

func TestParseHandSummary(t *testing.T) {
	handText := testHands

	summary, _ := parseHandSummary(handText)

	
	summaryWant := Summary{
		Pot:            0.36,
		Rake:           0.01,
		CommunityCards: []string{"Qc As 3d 2h"},
	}

	
	if !reflect.DeepEqual(summary, summaryWant) {
		t.Errorf("got %#v, but wanted %#v", summary, summaryWant)
	}
}

func TestParseMetadata(t *testing.T) {
	handText := testHands
	handTime, _ := time.Parse(time.DateTime, "2025-01-21 20:51:32")

	metadata, _ := parseMetaData(handText)

	metadataWant := Metadata{
		ID:   "254489598204",
		Date: handTime.Local(),
	}

	if metadata != metadataWant {
		t.Errorf("got %#v, but wanted %#v", metadata, metadataWant)
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
			got, err := actionPlayerNameFromText(c)

			if err != nil {
				t.Error("failed to parse name but should have")
			}

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
				got, _ := actionAmountFromText(c)
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
				_, err := actionAmountFromText(c)
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
Dealt to KavarzE [Ad Ac]
KavarzE: bets $2.33`)},
		}

		got, err := handsFromSessionFile(fileSystem, "zoom.txt")

		want := []Hand{
			{
				Metadata{"123", time.Time{}.Local()},
				[]Player{{Username: "KavarzE", Cards: "Ad Ac"}}, []Action{
					{"KavarzE", 1, Preflop, Bets, 2.33},
				},
				Summary{
					nil, 0, 0,
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
Dealt to KavarzE [Ad Ac]
KavarzE: bets $2.33`)

		got, _ := parseHands(handData)
		want := []Hand{
			{
				Metadata{
					"123", time.Time{}.Local(),
				},
				[]Player{{Username: "KavarzE", Cards: "Ad Ac"}}, []Action{{"KavarzE", 1, Preflop, Bets, 2.33}},
				Summary{
					nil, 0, 0,
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, wanted %#v", got, want)
		}
	})

	t.Run("Random non-hand data", func(t *testing.T) {
		handData := []byte(`Random non-hand data, whoops!`)

		_, err := parseHands(handData)

		if err == nil {
			t.Errorf("expected an error but didn't get one")
		}

		if !errors.Is(ErrNoHandID, errors.Unwrap(err[0])) {
			t.Errorf("expected a %v type of error but got a different one: %v", ErrNoHandID, err)
		}
	})

	t.Run("file with 3 hands, but one is corrupted", func(t *testing.T) {
		handData := []byte(brokenHands)
		hands, err := parseHands(handData)

		if len(hands) != 2 {
			// fmt.Printf("%#v\n\n %#v", hands, err)
			t.Errorf("wanted 2 hands and 1 error but got %#v hands", len(hands))
		}

		if len(err) != 1 {
			t.Errorf("wanted 1 error but got %v", len(err))
		}
	})
}

func TestHandIdFromText(t *testing.T) {
	handData := "Pokerstars Hand #6548679821301346841: Holdem don't care"

	got := handIDFromText(handData)
	want := "6548679821301346841"

	if got != want {
		t.Errorf("got %v wanted %v", got, want)
	}
}

func TestActionsFromText(t *testing.T) {
	// dummyActions := []Action{
	// 	{Player{"Kavarz", ""}, 2, Flop, Bets, 3},
	// 	{Player{"Burty", ""}, 3, Flop, Calls, 3},
	// }
	var dummyStreet = Flop
	order := 1

	handData := `Kavarz: bets $3`

	got, _, err := parseActionLine(handData, &dummyStreet, &order)
	if err != nil {
		t.Error(err)
	}

	want := Action{
		PlayerName: "Kavarz",
		Order:      2,
		Street:     Flop,
		ActionType: Bets,
		Amount:     3,
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

			got := dateTimeFromText(tt.test)
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

func TestPlayerCardsFromText(t *testing.T) {
	cases := []struct {
		test string
		want Player
	}{
		{`Seat 2: KavarzE (small blind) showed [Jc Js] and won ($5.03) with three of a kind, Jacks, and lost with three of a kind, Jacks`, Player{"KavarzE", "Jc Js"}},
		{`Seat 5: ThxWasOby3 showed [Ah Qd] and lost with high card Ace, and won ($5.02) with a flush, Ace high`, Player{"ThxWasOby3", "Ah Qd"}},
		{`Seat 6: KavarzE mucked [6s 6d]`, Player{"KavarzE", "6s 6d"}},
		{`Seat 5: ilbeback2017 showed [Tc Td] and won ($1.37) with a pair of Tens`, Player{"ilbeback2017", "Tc Td"}},
		{`Seat 1: acsy797 (button) mucked [Jd Ks]`, Player{"acsy797", "Jd Ks"}},
		{`Seat 1: KavarzE (big blind) collected ($0.04)`, Player{"KavarzE", ""}},
		{`Seat 4: VanillaLight (big blind) collected ($0.04)`, Player{"VanillaLight", ""}},
		{`Seat 4: JSIrony collected ($0.23)`, Player{"JSIrony", ""}},
		{`Seat 6: Imbastrol folded before Flop (didn't bet)`, Player{"Imbastrol", ""}},
	}

	for _, tt := range cases {
		t.Run(tt.test, func(t *testing.T) {

			var prefix string
			if strings.Contains(tt.test, "showed [") {
				prefix = "showed ["
			}
			if strings.Contains(tt.test, "mucked [") {
				prefix = "mucked ["
			}

			got := playerInfoFromText(tt.test, prefix)

			if got != tt.want {
				t.Errorf("got %v but we wanted %v", got, tt.want)
			}
		})
	}
}

func TestAmountFromText(t *testing.T) {

	t.Run("happy path pot", func(t *testing.T) {
		cases := []struct {
			test string
			want float64
		}{
			{"Total pot $0.94 | ", 0.94},
			{"Total pot $10.55 | Rake", 10.55},
			{"Total pot $2.36 | ", 2.36},
			{"Total pot $0.05 |", 0.05},
		}

		for _, tt := range cases {
			got, err := amountFromText(tt.test, potSizeSignifier)

			if err != nil {
				t.Error("expected nil error but got one")
			}

			if got != tt.want {
				t.Errorf("got %f wanted %f", got, tt.want)
			}
		}
	})

	t.Run("happy path rake", func(t *testing.T) {
		cases := []struct {
			test string
			want float64
		}{
			{"Rake $0.94\n", 0.94},
			{"Rake $10.55\n", 10.55},
			{"Rake $2.36\n", 2.36},
			{"| Rake $0.05\n", 0.05},
		}

		for _, tt := range cases {
			got, err := amountFromText(tt.test, rakeSizeSignifier)

			if err != nil {
				t.Errorf("expected nil error but got %v", err)
			}

			if got != tt.want {
				t.Errorf("got %f wanted %f", got, tt.want)
			}
		}
	})

	t.Run("failing non-float strings", func(t *testing.T) {
		cases := []struct {
			test   string
			result float64
		}{
			{"oh no there's no float value here", 0},
			{"Total pot $ unable to parsey", 0},
			{"Rake $ unparseable", 0},
		}

		for _, tt := range cases {
			got, err := amountFromText(tt.test, potSizeSignifier)

			if err == nil {
				t.Errorf("expected err but didn't get one")
			}

			if got != tt.result {
				t.Errorf("got %v, but wanted %v", got, tt.result)
			}
		}
	})

}

type failingFS struct{}

func (f failingFS) Open(name string) (fs.File, error) {
	return nil, errors.New("oh no i always fail")
}

const testHands string = `PokerStars Zoom Hand #254489598204:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:51:32 WET [2025/01/21 15:51:32 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: JDfq28 ($5.11 in chips)
Seat 2: kv_def ($11.57 in chips)
Seat 3: KavarzE ($5 in chips)
Seat 4: MGPN ($4.63 in chips)
Seat 5: ikin23 ($7.63 in chips)
Seat 6: honda589 ($5.38 in chips)
kv_def: posts small blind $0.02
KavarzE: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [3c 7c]
MGPN: folds
ikin23: folds
honda589: folds
JDfq28: raises $0.07 to $0.12
kv_def: calls $0.10
KavarzE: calls $0.07
*** FLOP *** [Qc As 3d]
kv_def: checks
KavarzE: checks
JDfq28: checks
*** TURN *** [Qc As 3d] [2h]
kv_def: bets $0.20
KavarzE: folds
JDfq28: folds
Uncalled bet ($0.20) returned to kv_def
kv_def collected $0.35 from pot
kv_def: doesn't show hand
*** SUMMARY ***
Total pot $0.36 | Rake $0.01
Board [Qc As 3d 2h]
Seat 1: JDfq28 (button) folded on the Turn
Seat 2: kv_def (small blind) collected ($0.35)
Seat 3: KavarzE (big blind) folded on the Turn
Seat 4: MGPN folded before Flop (didn't bet)
Seat 5: ikin23 folded before Flop (didn't bet)
Seat 6: honda589 folded before Flop (didn't bet)`

// const testHand string = `KavarzE: posts small blind $0.02
// getaddicted: posts big blind $0.05
// Mythic Max: folds
// Cl8rker: folds
// MGPN: raises $0.05 to $0.10
// SyraXmaX: folds
// KavarzE: folds
// getaddicted: folds`

const brokenHands string = `PokerStars Zoom Hand #254671589924:  Hold'em No Limit ($0.02/$0.05) - 2025/02/02 16:57:48 WET [2025/02/02 11:57:48 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: KavarzE ($5 in chips) 
Seat 2: Zwenni24 ($12.59 in chips) 
Seat 3: axf888 ($4.06 in chips) 
Seat 4: Mallorny ($5 in chips) 
Seat 5: bountykid99 ($6.91 in chips) 
Seat 6: vreska ($6.51 in chips) 
Zwenni24: posts small blind $0.02
axf888: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [6d 7c]
Mallorny: folds 
bountykid99: folds 
vreska: folds 
KavarzE: folds 
Zwenni24: raises $0.05 to $0.10
axf888: folds 
Uncalled bet ($0.05) returned to Zwenni24
Zwenni24 collected $0.10 from pot
Zwenni24: doesn't show hand 
*** SUMMARY ***
Total pot $0.10 | Rake $0 
Seat 1: KavarzE (button) folded before Flop (didn't bet)
Seat 2: Zwenni24 (small blind) collected ($0.10)
Seat 3: axf888 (big blind) folded before Flop
Seat 4: Mallorny folded before Flop (didn't bet)
Seat 5: bountykid99 folded before Flop (didn't bet)
Seat 6: vreska folded before Flop (didn't bet)



PokerStars Zoom Hand #254671591484:  Hold'em No Limit ($0.02/$0.05) - 2025/02/02 16:57:56 WET [2025/02/02 11:57:56 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: ikurah ($5 in chips) 
Seat 2: KavarzE ($5 in chips) 
Seat 3: Makitox ($5 in chips) 
Seat 4: Nadolf51 ($8.15 in chips) 
Seat 5: don caco 10 ($3.23 in chips) 
Seat 6: 22_Olga ($4.94 in chips) 
KavarzE: posts small blind $0.02
Makitox: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [Kh Qh]
Nadolf51: folds 
don caco 10: folds 
22_Olga: raises $0.10 to $0.15
ikurah: folds 
KavarzE: raises 0.45 to 0.60
Makitox: folds 
22_Olga: folds 
Uncalled bet ($0.45) returned to KavarzE
KavarzE collected $0.35 from pot
KavarzE: doesn't show hand 
*** SUMMARY ***
Total pot $0.35 | Rake $0 
Seat 1: ikurah (button) folded before Flop (didn't bet)
Seat 2: KavarzE (small blind) collected ($0.35)
Seat 3: Makitox (big blind) folded before Flop
Seat 4: Nadolf51 folded before Flop (didn't bet)
Seat 5: don caco 10 folded before Flop (didn't bet)
Seat 6: 22_Olga folded before Flop



PokerStars Zoom Hand #254671585485:  Hold'em No Limit ($0.02/$0.05) - 2025/02/02 16:57:25 WET [2025/02/02 11:57:25 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: ManeAlhekine ($5.25 in chips) 
Seat 2: TANEV78 ($5 in chips) 
Seat 3: KavarzE ($5 in chips) 
Seat 4: ewpd ($11.06 in chips) 
Seat 5: psbrets ($5.84 in chips) 
Seat 6: AQsuit ($5.11 in chips) 
TANEV78: posts small blind $0.02
KavarzE: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [7d 8c]
ewpd: folds 
psbrets: folds 
AQsuit: raises $0.07 to $0.12
ManeAlhekine has timed out
ManeAlhekine: folds 
TANEV78: calls $0.10
KavarzE: folds 
*** FLOP *** [Ac 6h 2h]
TANEV78: checks 
AQsuit: checks 
*** TURN *** [Ac 6h 2h] [6d]
TANEV78: checks 
AQsuit: checks 
*** RIVER *** [Ac 6h 2h 6d] [4c]
TANEV78: bets $0.10
AQsuit: folds 
Uncalled bet ($0.10) returned to TANEV78
TANEV78 collected $0.28 from pot
TANEV78: doesn't show hand 
*** SUMMARY ***
Total pot $0.29 | Rake $0.01 
Board [Ac 6h 2h 6d 4c]
Seat 1: ManeAlhekine (button) folded before Flop (didn't bet)
Seat 2: TANEV78 (small blind) collected ($0.28)
Seat 3: KavarzE (big blind) folded before Flop
Seat 4: ewpd folded before Flop (didn't bet)
Seat 5: psbrets folded before Flop (didn't bet)
Seat 6: AQsuit folded on the River`
