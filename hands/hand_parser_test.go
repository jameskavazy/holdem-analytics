package hands

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

func TestParseHandsAcceptance(t *testing.T) {
	t.Run("hand data is correctly parsed from a text file", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Wei III": {Data: []byte(cashGame2)},
		}
		file, _ := fileSystem.Open("Wei III")
		scanner := bufio.NewScanner(file)
		channel := make(chan handImport, 1)
		ok, scanErr := parseHands("Wei III ", scanner, channel)

		if !ok {
			t.Fatal("wanted ok=true from parseHands but got false")
		}
		if scanErr != nil {
			t.Errorf("wanted nil scanErr but got %v", scanErr)
		}

		handTime, _ := time.Parse(time.DateTime, "2025-01-19 12:38:55")

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "254446123323",
				Date:       handTime.Local(),
				ButtonSeat: 0,
			},
			Players: []Player{{"maximoIV", [2]Card{}, 1, 5.2}, {"dlourencobss", [2]Card{"8s", "9s"}, 2, 4.94}, {"KavarzE", [2]Card{"2s", "5d"}, 3, 5}, {"arsad725", [2]Card{}, 4, 5.49}, {"RE0309", [2]Card{}, 5, 4.63}, {"pernadao1599", [2]Card{"Jh", "Qc"}, 6, 3.43}},
			Actions: []Action{
				actionBuildHelper("dlourencobss", Posts, Preflop, 1, 0.02),
				actionBuildHelper("KavarzE", Posts, Preflop, 2, 0.05),
				actionBuildHelper("arsad725", Folds, Preflop, 3, 0),
				actionBuildHelper("RE0309", Calls, Preflop, 4, 0.05),
				actionBuildHelper("pernadao1599", Calls, Preflop, 5, 0.05),
				actionBuildHelper("maximoIV", Folds, Preflop, 6, 0),
				actionBuildHelper("dlourencobss", Calls, Preflop, 7, 0.03),
				actionBuildHelper("KavarzE", Checks, Preflop, 8, 0),
				actionBuildHelper("dlourencobss", Bets, Flop, 9, 0.10),
				actionBuildHelper("KavarzE", Folds, Flop, 10, 0),
				actionBuildHelper("RE0309", Folds, Flop, 11, 0),
				actionBuildHelper("pernadao1599", Calls, Flop, 12, 0.10),
				actionBuildHelper("dlourencobss", Bets, Turn, 13, 0.27),
				actionBuildHelper("pernadao1599", Calls, Turn, 14, 0.27),
				actionBuildHelper("dlourencobss", Checks, River, 15, 0),
				actionBuildHelper("pernadao1599", Checks, River, 16, 0),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{{
					Flop:  [3]Card{"2h", "Ts", "Jc"},
					Turn:  Card("3h"),
					River: Card("8c"),
				}, {}},
				Pot:  0.94,
				Rake: 0.05,
				Winners: []Winner{
					{"pernadao1599", 0.89},
				},
			},
		}

		assertHand(t, got.hand, want)
	})

	// t.Run("run it twice hand parse correctly", func(t *testing.T) {
	// 	fileSystem := fstest.MapFS{
	// 		"RIT": {Data: []byte(runItTwice)},
	// 	}
	// 	handHistory, _ := ExportHands(fileSystem)
	// 	handTime, _ := time.Parse(time.DateTime, "2025-01-29 16:30:35")

	// 	got := handHistory[0]
	// 	want := Hand{
	// 		Metadata: Metadata{
	// 			ID:   "254607988518",
	// 			Date: handTime.Local(),
	// 		},
	// 		Players: []Player{{"KavarzE", "Jc Js"}, {"TurivVB240492", ""}, {"RoMike2", ""}, {"hiroakin", ""}, {"ThxWasOby3", "Ah Qd"}, {"VLSALT", ""}},
	// 		Actions: []Action{
	// 			actionBuildHelper("KavarzE", Posts, Preflop, 1, 0.02),
	// 			actionBuildHelper("RoMike2", Posts, Preflop, 2, 0.05),
	// 			actionBuildHelper("hiroakin", Folds, Preflop, 3, 0.0),
	// 			actionBuildHelper("ThxWasOby3", Raises, Preflop, 4, 0.10),
	// 			actionBuildHelper("VLSALT", Folds, Preflop, 5, 0),
	// 			actionBuildHelper("TurivVB240492", Folds, Preflop, 6, 0),
	// 			actionBuildHelper("KavarzE", Raises, Preflop, 7, 0.45),
	// 			actionBuildHelper("RoMike2", Folds, Preflop, 8, 0),
	// 			actionBuildHelper("ThxWasOby3", Raises, Preflop, 9, 0.72),
	// 			actionBuildHelper("KavarzE", Calls, Preflop, 10, 0.72),
	// 			actionBuildHelper("KavarzE", Checks, Flop, 11, 0),
	// 			actionBuildHelper("ThxWasOby3", Checks, Flop, 12, 0),
	// 			actionBuildHelper("KavarzE", Bets, Turn, 13, 1.81),
	// 			actionBuildHelper("ThxWasOby3", Raises, Turn, 14, 2.09),
	// 			actionBuildHelper("KavarzE", Calls, Turn, 15, 2.09),
	// 		},
	// 		Summary: Summary{
	// 			CommunityCards: []string{"7d 2h 8h Jh 3d", "7d 2h 8h Jh Qh"},
	// 			Pot:            10.49,
	// 			Rake:           0.44,
	// 		},
	// 	}
	// 	assertHand(t, got, want)
	// })
}

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
		Pot:  0.36,
		Rake: 0.01,
		CommunityCards: [2]CommunityCards{
			{[3]Card{"Qc", "As", "3d"},
				Card("2h"),
				Card(""),
			},
			{}},
		Winners: []Winner{
			{"kv_def", 0.35},
		},
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

	t.Run("player bets all-in", func(t *testing.T) {
		cases := map[string]float64{
			"Krawicz: bets $1.27 and is all-in":  1.27,
			"windy886: bets $0.76 and is all-in": 0.76,
		}

		for c, want := range cases {
			t.Run(c, func(t *testing.T) {
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

func TestExtractHandsFromFile(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"zoom.txt": {Data: []byte(`PokerStars Hand #123: blah blah
Seat 1: test ($6000 in chips)
Seat 2: KavarzE ($3000 in chips)
Dealt to KavarzE [Ad Ac]
KavarzE: bets $2.33`)},
		}

		handChan := make(chan handImport, 1)

		var result bool
		var fsErr error
		result, fsErr = extractHandsFromFile(fileSystem, "zoom.txt", handChan)

		got := <-handChan

		want := handImport{
			"zoom.txt",
			Hand{
				Metadata{"123", time.Time{}.Local(), 0},
				[]Player{
					{Username: "KavarzE"},
					{Cards: [2]Card{"Ad", "Ac"}},
					{Seat: 2},
					{ChipCount: 3000},
				},
				[]Action{
					{"KavarzE", 1, Preflop, Bets, 2.33},
				},
				Summary{
					[2]CommunityCards{}, 0, 0, 0, []Winner{},
				},
			},
			nil,
			false,
		}

		if got.handErr != nil {
			t.Fatal("got an error but didn't expect one: ", got.handErr)
		}

		if got.hand.Summary.CommunityCards != [2]CommunityCards{} {
			t.Errorf("Summary.Community Cards: got %#v wanted %#v", got, want)
		}

		if got.hand.Metadata.ID != want.hand.Metadata.ID {
			t.Errorf("Metadata ID: got %#v wanted %#v", got, want)
		}

		if !result {
			t.Fatal("got false result, expected true")
		}

		if fsErr != nil {
			t.Fatal("got a filesystem error but didn't expect one: ", fsErr)
		}
	})

	t.Run("error pathway", func(t *testing.T) {
		fileSystem := failingFS{}
		handChan := make(chan handImport, 10000)
		ok, fsErr := extractHandsFromFile(fileSystem, "zoom.txt", handChan)

		if fsErr == nil {
			t.Fatal("expected an fsError but didn't get one!")
		}

		if ok {
			t.Fatal("expected ok=false but was true!")
		}

	})
}

func TestParseHandData(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		filename := "test-file.txt"
		handData := []byte(`PokerStars Hand #123: blah blah
Seat 1: test ($6000 in chips)
Seat 2: test2 ($3000 in chips)
Dealt to test [Ad Ac]
test: bets $2.33`)

		bytesReader := bytes.NewReader(handData)
		scanner := bufio.NewScanner(bytesReader)

		handCh := make(chan handImport, 10000)

		ok, scanErr := parseHands(filename, scanner, handCh)

		got := <-handCh

		if !ok {
			t.Fatal("expected ok to be true but was false!")
		}

		if scanErr != nil {
			t.Errorf("expected nil error but got %#v", scanErr)
		}

		want := handImport{
			filename,
			Hand{
				Metadata{"123", time.Time{}.Local(), 0},
				[]Player{
					{Username: "test", Cards: [2]Card{"Ad", "Ac"}, Seat: 1, ChipCount: 6000},
					{Username: "test2", Cards: [2]Card{"", ""}, Seat: 2, ChipCount: 3000}},
				[]Action{{"test", 1, Preflop, Bets, 2.33}},
				Summary{[2]CommunityCards{}, 0, 0, 0, nil},
			},
			nil,
			false,
		}

		if !reflect.DeepEqual(got.hand, want.hand) {
			t.Errorf("got %#v, wanted %#v", got, want)
		}
	})

	t.Run("Random non-hand data", func(t *testing.T) {
		filename := "filename.txt"
		handData := []byte(`Random non-hand data, whoops!`)
		bytesReader := bytes.NewReader(handData)
		scanner := bufio.NewScanner(bytesReader)

		handChan := make(chan handImport, 10000)

		ok, scanErr := parseHands(filename, scanner, handChan)

		if !ok || scanErr != nil {
			t.Errorf("wanted non-nil error and ok=true, but got error: %#v, ok: %#v ", scanErr, ok)
		}

		got := <-handChan

		if got.handErr == nil {
			t.Errorf("expected an error but didn't get one")
		}

	})

	t.Run("file with 3 hands, but one is corrupted", func(t *testing.T) {
		filename := "filename.txt"
		handData := []byte(brokenHands)
		bytesReader := bytes.NewReader(handData)
		scanner := bufio.NewScanner(bytesReader)

		handChan := make(chan handImport, 10000)

		var wg sync.WaitGroup
		var ok bool
		var scanErr error
		wg.Add(1)

		go func() {
			defer wg.Done()
			ok, scanErr = parseHands(filename, scanner, handChan)
		}()

		wg.Wait()
		close(handChan)

		if !ok || scanErr != nil {
			t.Errorf("wanted non-nil error and not ok, but got error: %#v, ok: %#v ", scanErr, ok)
		}

		var hands []Hand
		var handErrs []error

		for h := range handChan {
			hands = append(hands, h.hand)
			handErrs = append(handErrs, h.handErr)
		}

		var handCount int

		for _, h := range hands {
			if !reflect.DeepEqual(h, Hand{}) {
				handCount++
			}
		}
		if len(hands) != 3 {
			t.Errorf("wanted 3 hands and 1 error but got %#v hands", len(hands))
		}

		var errCount int
		for _, e := range handErrs {
			if e != nil {
				errCount++
			}
		}

		if errCount != 1 {
			t.Errorf("wanted 1 error but got %d, Hand Err Slice: %#v", errCount, handErrs)
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

func TestSeatIntFromText(t *testing.T) {
	cases := []struct {
		test string
		want int64
	}{
		{`Seat 6: KavarzE ($1.97 in chips) `, 6},
		{`Seat 5: NicSt4r ($2 in chips) `, 5},
		{`Seat 1: superRODRI ($1.18 in chips)`, 1},
	}
	for _, tt := range cases {
		t.Run(tt.test, func(t *testing.T) {
			got, err := seatIntFromText(tt.test)

			if got != tt.want {
				t.Errorf("got %d but we wanted %d", got, tt.want)
			}

			if err != nil {
				t.Fatal("got an error but didn't expected one")
			}
		})
	}
}

func TestCommunityCardsFromText(t *testing.T) {
	handText := testHands
	got := communityCardsFromText(handText, boardSignifier)

	if got.Flop != [3]Card{"Qc", "As", "3d"} {
		t.Errorf("wanted %#v community cards but got %#v", [3]Card{"Qc", "As", "3d"}, got.Flop)
	}

	if got.Turn != Card("2h") {
		t.Errorf("wanted %v community cards but got %v", Card("2h"), got.Turn)
	}
}

func TestPlayerCardsFromText(t *testing.T) {
	cases := []struct {
		test string
		want Player
	}{
		{`Seat 2: KavarzE (small blind) showed [Jc Js] and won ($5.03) with three of a kind, Jacks, and lost with three of a kind, Jacks`, Player{"KavarzE", [2]Card{"Jc", "Js"}, 0, 0}},
		{`Seat 1: acsy797 (button) mucked [Jd Ks]`, Player{"acsy797", [2]Card{"Jd", "Ks"}, 0, 0}},
		// {`Seat 1: KavarzE (big blind) collected ($0.04)`, Player{"KavarzE", ""}},
		// {`Seat 4: VanillaLight (big blind) collected ($0.04)`, Player{"VanillaLight", ""}},
		// {`Seat 4: JSIrony collected ($0.23)`, Player{"JSIrony", ""}},
		// {`Seat 6: Imbastrol folded before Flop (didn't bet)`, Player{"Imbastrol", ""}},
		{`Dealt to KavarzE [Js 5c]`, Player{"KavarzE", [2]Card{"Js", "5c"}, 0, 0}},
		{`Seat 6: KavarzE ($1.97 in chips) `, Player{"KavarzE", [2]Card{}, 6, 1.97}},
	}

	for _, tt := range cases {
		t.Run(tt.test, func(t *testing.T) {

			// var prefix string
			// if strings.Contains(tt.test, "showed [") {
			// 	prefix = "showed ["
			// }
			// if strings.Contains(tt.test, "mucked [") {
			// 	prefix = "mucked ["
			// }
			// if strings.Contains(tt.test, "Dealt to") {
			// 	prefix = "["
			// }
			// if strings.Contains(tt.test, "Seat ") {

			// }

			got, _, _ := parsePlayer(tt.test)

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

func TestUpdateOrAppendPlayer(t *testing.T) {
	players := map[string]*Player{
		"KavarzE": {"KavarzE", [2]Card{"", ""}, 1, 6.00},
		"Javormy": {"Javormy", [2]Card{"", ""}, 2, 33.00},
		"noob":    {"noob", [2]Card{"", ""}, 3, 4.00},
	}

	updateOrAddPlayer(
		players,
		Player{"KavarzE", [2]Card{"Ac", "Ad"}, 1, 6.00})

	if len(players) != 3 {
		t.Errorf("expected player length of 3 but got %v", len(players))
	}

	for _, p := range players {
		if p.Username == "KavarzE" && p.Cards != [2]Card{"Ac", "Ad"} {
			t.Errorf("wanted updated cards of %v but got %v", [2]Card{"Ac", "Ad"}, p.Cards)
		}
	}
}

func TestConvertToSlice(t *testing.T) {
	players := map[string]*Player{
		"KavarzE": {"KavarzE", [2]Card{"", ""}, 1, 6.00},
		"Javormy": {"Javormy", [2]Card{"", ""}, 3, 33.00},
		"noob":    {"noob", [2]Card{"", ""}, 2, 4.00},
	}

	got := convertToSlice(players)

	if got[1].Username != "noob" {
		t.Errorf("wanted username: 'noob' in position 1 in slice, but got username of %v", got[1].Username)
	}

	if got[2].Username != "Javormy" {
		t.Errorf("wanted username: 'Javormy' in position 2 in slice, but got username of %v", got[2].Username)
	}
}

func TestWinnerFromHandText(t *testing.T) {

	cases := []struct {
		test string
		want Winner
	}{

		{"Seat 2: kv_def (small blind) collected ($0.35)", Winner{"kv_def", 0.35}},
		{"Seat 4: lukebartlett showed [Qd Kc] and won ($2.99) with two pair, Kings and Queens", Winner{"lukebartlett", 2.99}},
	}

	for _, tt := range cases {
		got, err := winnerFromLine(tt.test)
		if err != nil {
			t.Fatalf("got an error but got: %v", err)
		}

		if got.PlayerName != tt.want.PlayerName {
			t.Errorf("test case: '%v':\n wanted %v as winning player but got %v ", tt.test, tt.want.PlayerName, got.PlayerName)
		}

		if got.Amount != tt.want.Amount {
			t.Errorf("test case: '%v':\n wanted %v as winning amount but got %v ", tt.test, tt.want.Amount, got.Amount)
		}
	}

}

type failingFS struct{}

func (f failingFS) Open(_ string) (fs.File, error) {
	return nil, errors.New("oh no i always fail")
}

func assertHand(t *testing.T, got, want Hand) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %#v, \n\ngot %#v", want, got)
	}
}

func actionBuildHelper(playerName string, actionType ActionType, street Street, order int, amount float64) Action {
	return Action{
		PlayerName: playerName,
		ActionType: actionType,
		Street:     street,
		Order:      order,
		Amount:     amount,
	}
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

const cashGame2 string = `PokerStars Hand #254446123323:  Hold'em No Limit ($0.02/$0.05 USD) - 2025/01/19 12:38:55 WET [2025/01/19 7:38:55 ET]
Table 'Wei III' 6-max Seat #1 is the button
Seat 1: maximoIV ($5.20 in chips)
Seat 2: dlourencobss ($4.94 in chips)
Seat 3: KavarzE ($5 in chips)
Seat 4: arsad725 ($5.49 in chips)
Seat 5: RE0309 ($4.63 in chips)
Seat 6: pernadao1599 ($3.43 in chips)
dlourencobss: posts small blind $0.02
KavarzE: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [2s 5d]
arsad725: folds
RE0309: calls $0.05
pernadao1599: calls $0.05
maximoIV: folds
dlourencobss: calls $0.03
KavarzE: checks
*** FLOP *** [2h Ts Jc]
dlourencobss: bets $0.10
KavarzE: folds
RE0309: folds
pernadao1599: calls $0.10
*** TURN *** [2h Ts Jc] [3h]
dlourencobss: bets $0.27
pernadao1599: calls $0.27
*** RIVER *** [2h Ts Jc 3h] [8c]
dlourencobss: checks
pernadao1599: checks
*** SHOW DOWN ***
dlourencobss: shows [8s 9s] (a pair of Eights)
pernadao1599: shows [Jh Qc] (a pair of Jacks)
pernadao1599 collected $0.89 from pot
*** SUMMARY ***
Total pot $0.94 | Rake $0.05
Board [2h Ts Jc 3h 8c]
Seat 1: maximoIV (button) folded before Flop (didn't bet)
Seat 2: dlourencobss (small blind) showed [8s 9s] and lost with a pair of Eights
Seat 3: KavarzE (big blind) folded on the Flop
Seat 4: arsad725 folded before Flop (didn't bet)
Seat 5: RE0309 folded on the Flop
Seat 6: pernadao1599 showed [Jh Qc] and won ($0.89) with a pair of Jacks`

const multipleWinnersHand string = `PokerStars Zoom Hand #257507021156:  Hold'em No Limit ($0.01/$0.02) - 2025/08/27 17:58:57 WET [2025/08/27 12:58:57 ET]
Table 'Halley' 6-max Seat #1 is the button
Seat 1: KavarzE ($2 in chips) 
Seat 2: gepard35 ($2.83 in chips) 
Seat 3: Javis1311 ($1 in chips) 
Seat 4: ricardo_riro ($2 in chips) 
Seat 5: ferchaPok ($2.04 in chips) 
Seat 6: ChipInvadr ($5.53 in chips) 
gepard35: posts small blind $0.01
Javis1311: posts big blind $0.02
*** HOLE CARDS ***
Dealt to KavarzE [6d Th]
ricardo_riro: folds 
ferchaPok: folds 
ChipInvadr: folds 
KavarzE: folds 
gepard35: raises $0.04 to $0.06
Javis1311: calls $0.04
*** FLOP *** [Jd Ah 8s]
gepard35: bets $0.05
Javis1311: calls $0.05
*** TURN *** [Jd Ah 8s] [Ts]
gepard35: bets $0.08
Javis1311: calls $0.08
*** RIVER *** [Jd Ah 8s Ts] [8c]
gepard35: bets $0.28
Javis1311: raises $0.53 to $0.81 and is all-in
gepard35: calls $0.53
*** SHOW DOWN ***
Javis1311: shows [Ad Td] (two pair, Aces and Tens)
gepard35: shows [Ac Tc] (two pair, Aces and Tens)
gepard35 collected $0.97 from pot
Javis1311 collected $0.96 from pot
*** SUMMARY ***
Total pot $2 | Rake $0.07 
Board [Jd Ah 8s Ts 8c]
Seat 1: KavarzE (button) folded before Flop (didn't bet)
Seat 2: gepard35 (small blind) showed [Ac Tc] and won ($0.97) with two pair, Aces and Tens
Seat 3: Javis1311 (big blind) showed [Ad Td] and won ($0.96) with two pair, Aces and Tens
Seat 4: ricardo_riro folded before Flop (didn't bet)
Seat 5: ferchaPok folded before Flop (didn't bet)
Seat 6: ChipInvadr folded before Flop (didn't bet)`
