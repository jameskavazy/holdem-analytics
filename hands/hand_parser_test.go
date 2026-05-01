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
	loc, _ := time.LoadLocation("America/New_York")
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

		etTimeStr := "2025-01-19 7:38:55"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "254446123323",
				Date:       wantTime,
				ButtonSeat: 1,
			},
			Players: []Player{{"maximoIV", [2]Card{}, 1, 5.2}, {"dlourencobss", [2]Card{"8s", "9s"}, 2, 4.94}, {"KavarzE", [2]Card{"2s", "5d"}, 3, 5}, {"arsad725", [2]Card{}, 4, 5.49}, {"RE0309", [2]Card{}, 5, 4.63}, {"pernadao1599", [2]Card{"Jh", "Qc"}, 6, 3.43}},
			Actions: []Action{
				actionBuildHelper("dlourencobss", ActionPost, Preflop, 1, 0.02),
				actionBuildHelper("KavarzE", ActionPost, Preflop, 2, 0.05),
				actionBuildHelper("arsad725", ActionFold, Preflop, 3, 0),
				actionBuildHelper("RE0309", ActionCall, Preflop, 4, 0.05),
				actionBuildHelper("pernadao1599", ActionCall, Preflop, 5, 0.05),
				actionBuildHelper("maximoIV", ActionFold, Preflop, 6, 0),
				actionBuildHelper("dlourencobss", ActionCall, Preflop, 7, 0.03),
				actionBuildHelper("KavarzE", ActionCheck, Preflop, 8, 0),
				actionBuildHelper("dlourencobss", ActionBet, Flop, 9, 0.10),
				actionBuildHelper("KavarzE", ActionFold, Flop, 10, 0),
				actionBuildHelper("RE0309", ActionFold, Flop, 11, 0),
				actionBuildHelper("pernadao1599", ActionCall, Flop, 12, 0.10),
				actionBuildHelper("dlourencobss", ActionBet, Turn, 13, 0.27),
				actionBuildHelper("pernadao1599", ActionCall, Turn, 14, 0.27),
				actionBuildHelper("dlourencobss", ActionCheck, River, 15, 0),
				actionBuildHelper("pernadao1599", ActionCheck, River, 16, 0),
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
					{"pernadao1599", 0.89, 1},
				},
			},
		}

		assertHand(t, got.hand, want)
	})

	t.Run("hand data is correctly for an uncalled bet instance", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Wei III": {Data: []byte(uncalledBetHand)},
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

		etTimeStr := "2025-08-27 13:30:17"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "257507385322",
				Date:       wantTime,
				ButtonSeat: 1,
			},
			Players: []Player{{"TSCardinals", [2]Card{}, 1, 2.02}, {"Jimmey54", [2]Card{}, 2, 2.21}, {"nm8800", [2]Card{}, 3, 2.31}, {"Chewbacca97", [2]Card{}, 4, 1.08}, {"KavarzE", [2]Card{"8s", "As"}, 5, 2.08}, {"haeorm", [2]Card{}, 6, 6.26}},
			Actions: []Action{
				actionBuildHelper("Jimmey54", ActionPost, Preflop, 1, 0.01),
				actionBuildHelper("nm8800", ActionPost, Preflop, 2, 0.02),
				actionBuildHelper("Chewbacca97", ActionFold, Preflop, 3, 0),
				actionBuildHelper("KavarzE", ActionRaise, Preflop, 4, 0.04),
				actionBuildHelper("haeorm", ActionFold, Preflop, 5, 0),
				actionBuildHelper("TSCardinals", ActionCall, Preflop, 6, 0.06),
				actionBuildHelper("Jimmey54", ActionFold, Preflop, 7, 0),
				actionBuildHelper("nm8800", ActionFold, Preflop, 8, 0),
				actionBuildHelper("KavarzE", ActionBet, Flop, 9, 0.04),
				actionBuildHelper("TSCardinals", ActionCall, Flop, 10, 0.04),
				actionBuildHelper("KavarzE", ActionCheck, Turn, 11, 0),
				actionBuildHelper("TSCardinals", ActionBet, Turn, 12, 0.17),
				actionBuildHelper("KavarzE", ActionFold, Turn, 13, 0.00),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{{
					Flop:  [3]Card{"Tc", "4h", "6h"},
					Turn:  Card("5c"),
					River: Card(""),
				}, {}},
				Pot:  0.23,
				Rake: 0.01,
				Winners: []Winner{
					{"TSCardinals", 0.22, 1},
				},
			},
		}

		assertHand(t, got.hand, want)
	})

	t.Run("run it twice hand parse correctly", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"RIT": {Data: []byte(runItTwice)},
		}
		file, _ := fileSystem.Open("RIT")
		scanner := bufio.NewScanner(file)
		channel := make(chan handImport, 1)

		ok, _ := parseHands("RIT", scanner, channel)

		if !ok {
			t.Fatal("wanted parseHands to be ok=true but got false")
		}

		etTimeStr := "2025-01-29 11:30:35"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel
		want := Hand{
			Metadata: Metadata{
				ID:         "254607988518",
				Date:       wantTime.UTC(),
				ButtonSeat: 1,
			},
			Players: []Player{
				{"TurivVB240492", [2]Card{}, 1, 1.94},
				{"KavarzE", [2]Card{"Jc", "Js"}, 2, 15.14},
				{"RoMike2", [2]Card{}, 3, 5.07},
				{"hiroakin", [2]Card{}, 4, 5},
				{"ThxWasOby3", [2]Card{"Ah", "Qd"}, 5, 5.22},
				{"VLSALT", [2]Card{}, 6, 5},
			},
			Actions: []Action{
				actionBuildHelper("KavarzE", ActionPost, Preflop, 1, 0.02),
				actionBuildHelper("RoMike2", ActionPost, Preflop, 2, 0.05),
				actionBuildHelper("hiroakin", ActionFold, Preflop, 3, 0.0),
				actionBuildHelper("ThxWasOby3", ActionRaise, Preflop, 4, 0.10),
				actionBuildHelper("VLSALT", ActionFold, Preflop, 5, 0),
				actionBuildHelper("TurivVB240492", ActionFold, Preflop, 6, 0),
				actionBuildHelper("KavarzE", ActionRaise, Preflop, 7, 0.45),
				actionBuildHelper("RoMike2", ActionFold, Preflop, 8, 0),
				actionBuildHelper("ThxWasOby3", ActionRaise, Preflop, 9, 0.72),
				actionBuildHelper("KavarzE", ActionCall, Preflop, 10, 0.72),
				actionBuildHelper("KavarzE", ActionCheck, Flop, 11, 0),
				actionBuildHelper("ThxWasOby3", ActionCheck, Flop, 12, 0),
				actionBuildHelper("KavarzE", ActionBet, Turn, 13, 1.81),
				actionBuildHelper("ThxWasOby3", ActionRaise, Turn, 14, 2.09),
				actionBuildHelper("KavarzE", ActionCall, Turn, 15, 2.09),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{
					{Flop: [3]Card{"7d", "2h", "8h"},
						Turn:  Card("Jh"),
						River: Card("3d")},
					{
						Flop:  [3]Card{"7d", "2h", "8h"},
						Turn:  Card("Jh"),
						River: Card("Qh"),
					}},
				Pot:  10.49,
				Rake: 0.44,
				Winners: []Winner{
					{PlayerName: "KavarzE", Amount: 5.03, Board: 1},
					{PlayerName: "ThxWasOby3", Amount: 5.02, Board: 2},
				},
			},
		}
		assertHand(t, got.hand, want)
	})

	t.Run("multiple winners (split pot) rio showdown", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Halley": {Data: []byte(multipleWinnersHand)},
		}
		file, _ := fileSystem.Open("Halley")
		scanner := bufio.NewScanner(file)
		channel := make(chan handImport, 1)
		ok, scanErr := parseHands("Halley", scanner, channel)

		if !ok {
			t.Fatal("wanted ok=true from parseHands but got false")
		}
		if scanErr != nil {
			t.Errorf("wanted nil scanErr but got %v", scanErr)
		}

		etTimeStr := "2025-08-27 12:58:57"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "257507021156",
				Date:       wantTime,
				ButtonSeat: 1,
			},
			Players: []Player{
				{"KavarzE", [2]Card{"6d", "Th"}, 1, 2},
				{"gepard35", [2]Card{"Ac", "Tc"}, 2, 2.83},
				{"Javis1311", [2]Card{"Ad", "Td"}, 3, 1},
				{"ricardo_riro", [2]Card{}, 4, 2},
				{"ferchaPok", [2]Card{}, 5, 2.04},
				{"ChipInvadr", [2]Card{}, 6, 5.53},
			},
			Actions: []Action{
				actionBuildHelper("gepard35", ActionPost, Preflop, 1, 0.01),
				actionBuildHelper("Javis1311", ActionPost, Preflop, 2, 0.02),
				actionBuildHelper("ricardo_riro", ActionFold, Preflop, 3, 0),
				actionBuildHelper("ferchaPok", ActionFold, Preflop, 4, 0),
				actionBuildHelper("ChipInvadr", ActionFold, Preflop, 5, 0),
				actionBuildHelper("KavarzE", ActionFold, Preflop, 6, 0),
				actionBuildHelper("gepard35", ActionRaise, Preflop, 7, 0.04),
				actionBuildHelper("Javis1311", ActionCall, Preflop, 8, 0.04),
				actionBuildHelper("gepard35", ActionBet, Flop, 9, 0.05),
				actionBuildHelper("Javis1311", ActionCall, Flop, 10, 0.05),
				actionBuildHelper("gepard35", ActionBet, Turn, 11, 0.08),
				actionBuildHelper("Javis1311", ActionCall, Turn, 12, 0.08),
				actionBuildHelper("gepard35", ActionBet, River, 13, 0.28),
				actionBuildHelper("Javis1311", ActionRaise, River, 14, 0.53),
				actionBuildHelper("gepard35", ActionCall, River, 15, 0.53),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{{
					Flop:  [3]Card{"Jd", "Ah", "8s"},
					Turn:  Card("Ts"),
					River: Card("8c"),
				}, {}},
				Pot:  2.0,
				Rake: 0.07,
				Winners: []Winner{
					{"gepard35", 0.97, 1},
					{"Javis1311", 0.96, 1},
				},
			},
		}

		assertHand(t, got.hand, want)
	})

	t.Run("rit player won both boards", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Donati": {Data: []byte(runItTwicePlayerWonBothBoards)},
		}
		file, _ := fileSystem.Open("Donati")
		scanner := bufio.NewScanner(file)
		channel := make(chan handImport, 1)
		ok, scanErr := parseHands("Donati", scanner, channel)

		if !ok {
			t.Fatal("wanted ok=true from parseHands but got false")
		}
		if scanErr != nil {
			t.Errorf("wanted nil scanErr but got %v", scanErr)
		}

		etTimeStr := "2025-01-19 11:36:38"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "254449744546",
				Date:       wantTime,
				ButtonSeat: 1,
			},
			Players: []Player{
				{"AsmAngAmAngo", [2]Card{}, 1, 6.95},
				{"loto_insane", [2]Card{}, 2, 5},
				{"KavarzE", [2]Card{"As", "Jc"}, 3, 7.11},
				{"Braghinn", [2]Card{}, 4, 5.72},
				{"R.S.P747", [2]Card{}, 5, 5.51},
				{"Gatzin", [2]Card{"Qh", "Jh"}, 6, 6.88},
			},
			Actions: []Action{
				actionBuildHelper("loto_insane", ActionPost, Preflop, 1, 0.02),
				actionBuildHelper("KavarzE", ActionPost, Preflop, 2, 0.05),
				actionBuildHelper("Braghinn", ActionRaise, Preflop, 3, 0.06),
				actionBuildHelper("R.S.P747", ActionFold, Preflop, 4, 0),
				actionBuildHelper("Gatzin", ActionCall, Preflop, 5, 0.11),
				actionBuildHelper("AsmAngAmAngo", ActionFold, Preflop, 6, 0),
				actionBuildHelper("loto_insane", ActionCall, Preflop, 7, 0.09),
				actionBuildHelper("KavarzE", ActionRaise, Preflop, 8, 0.89),
				actionBuildHelper("Braghinn", ActionFold, Preflop, 9, 0),
				actionBuildHelper("Gatzin", ActionCall, Preflop, 10, 0.89),
				actionBuildHelper("loto_insane", ActionFold, Preflop, 11, 0),
				actionBuildHelper("KavarzE", ActionBet, Flop, 12, 1.60),
				actionBuildHelper("Gatzin", ActionCall, Flop, 13, 1.60),
				actionBuildHelper("KavarzE", ActionBet, Turn, 14, 4.51),
				actionBuildHelper("Gatzin", ActionCall, Turn, 15, 4.28),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{
					{
						Flop:  [3]Card{"Js", "7s", "8c"},
						Turn:  Card("6h"),
						River: Card("6d"),
					},
					{
						Flop:  [3]Card{"Js", "7s", "8c"},
						Turn:  Card("6h"),
						River: Card("Ks"),
					},
				},
				Pot:  13.98,
				Rake: 0.58,
				Winners: []Winner{
					{"KavarzE", 6.70, 1},
					{"KavarzE", 6.70, 2},
				},
			},
		}

		assertHand(t, got.hand, want)
	})

	t.Run("all folded before flop, board 0 winner", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Donati": {Data: []byte(allFoldedBeforeFlop)},
		}
		file, _ := fileSystem.Open("Donati")
		scanner := bufio.NewScanner(file)
		channel := make(chan handImport, 1)
		ok, scanErr := parseHands("Donati", scanner, channel)

		if !ok {
			t.Fatal("wanted ok=true from parseHands but got false")
		}
		if scanErr != nil {
			t.Errorf("wanted nil scanErr but got %v", scanErr)
		}

		etTimeStr := "2025-01-30 14:50:17"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "254626485418",
				Date:       wantTime,
				ButtonSeat: 1,
			},
			Players: []Player{
				{"OoJohnStevensoO", [2]Card{}, 1, 6.24},
				{"bk4crs", [2]Card{}, 2, 9.22},
				{"KavarzE", [2]Card{"Qd", "5c"}, 3, 5},
				{"FabuTK", [2]Card{}, 4, 4.35},
				{"getaddicted", [2]Card{}, 5, 6.59},
				{"ilbeback2017", [2]Card{}, 6, 19.69},
			},
			Actions: []Action{
				actionBuildHelper("bk4crs", ActionPost, Preflop, 1, 0.02),
				actionBuildHelper("KavarzE", ActionPost, Preflop, 2, 0.05),
				actionBuildHelper("FabuTK", ActionFold, Preflop, 3, 0),
				actionBuildHelper("getaddicted", ActionFold, Preflop, 4, 0),
				actionBuildHelper("ilbeback2017", ActionFold, Preflop, 5, 0),
				actionBuildHelper("OoJohnStevensoO", ActionRaise, Preflop, 6, 0.06),
				actionBuildHelper("bk4crs", ActionFold, Preflop, 7, 0),
				actionBuildHelper("KavarzE", ActionFold, Preflop, 8, 0),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{{}, {}},
				Pot:            0.12,
				Rake:           0,
				Winners: []Winner{
					{"OoJohnStevensoO", 0.12, 0},
				},
			},
		}

		assertHand(t, got.hand, want)
	})

	t.Run("rit edge case, split pot on second board", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Donati": {Data: []byte(ritEdgeCaseHand)},
		}
		file, _ := fileSystem.Open("Donati")
		scanner := bufio.NewScanner(file)
		channel := make(chan handImport, 1)
		ok, scanErr := parseHands("Donati", scanner, channel)

		if !ok {
			t.Fatal("wanted ok=true from parseHands but got false")
		}
		if scanErr != nil {
			t.Errorf("wanted nil scanErr but got %v", scanErr)
		}

		etTimeStr := "2025-01-30 14:51:09"
		parsedEt, _ := time.ParseInLocation(time.DateTime, etTimeStr, loc)
		wantTime := parsedEt.UTC()

		got := <-channel

		want := Hand{
			Metadata: Metadata{
				ID:         "254626500457",
				Date:       wantTime,
				ButtonSeat: 1,
			},
			Players: []Player{
				{"Zutuzutu_90", [2]Card{"Tc", "9c"}, 1, 7.31},
				{"KavarzE", [2]Card{"9s", "Ks"}, 2, 5},
				{"darchas", [2]Card{}, 3, 5},
				{"soyjuliansito", [2]Card{}, 4, 5.03},
				{"SpieWNogach", [2]Card{}, 5, 5.07},
				{"Trogloditapubg", [2]Card{}, 6, 4.75},
			},
			Actions: []Action{
				actionBuildHelper("KavarzE", ActionPost, Preflop, 1, 0.02),
				actionBuildHelper("darchas", ActionPost, Preflop, 2, 0.05),
				actionBuildHelper("soyjuliansito", ActionFold, Preflop, 3, 0),
				actionBuildHelper("SpieWNogach", ActionFold, Preflop, 4, 0),
				actionBuildHelper("Trogloditapubg", ActionFold, Preflop, 5, 0),
				actionBuildHelper("Zutuzutu_90", ActionRaise, Preflop, 6, 0.07),
				actionBuildHelper("KavarzE", ActionRaise, Preflop, 7, 0.33),
				actionBuildHelper("darchas", ActionFold, Preflop, 8, 0),
				actionBuildHelper("Zutuzutu_90", ActionCall, Preflop, 9, 0.33),
				actionBuildHelper("KavarzE", ActionBet, Flop, 10, 0.30),
				actionBuildHelper("Zutuzutu_90", ActionRaise, Flop, 11, 0.45),
				actionBuildHelper("KavarzE", ActionCall, Flop, 12, 0.45),
				actionBuildHelper("KavarzE", ActionBet, Turn, 13, 3.80),
				actionBuildHelper("Zutuzutu_90", ActionCall, Turn, 14, 3.80),
			},
			Summary: Summary{
				CommunityCards: [2]CommunityCards{
					{
						Flop:  [3]Card{"Ts", "2d", "8s"},
						Turn:  Card("7h"),
						River: Card("Kh"),
					},
					{
						Flop:  [3]Card{"Ts", "2d", "8s"},
						Turn:  Card("7h"),
						River: Card("6d"),
					},
				},
				Pot:  10.05,
				Rake: 0.45,
				Winners: []Winner{
					{"KavarzE", 4.82, 1},
					{"KavarzE", 2.41, 2},
					{"Zutuzutu_90", 2.37, 2},
				},
			},
		}

		assertHand(t, got.hand, want)
	})
}

func TestActionTypeFromText(t *testing.T) {
	cases := map[string]ActionType{
		"kv_def: posts small blind $0.02": ActionPost,
		"KavarzE: posts big blind $0.05":  ActionPost,
		"arsad725: folds":                 ActionFold,
		"RE0309: calls $0.05":             ActionCall,
		"pernadao1599: calls $0.05":       ActionCall,
		"maximoIV: folds ":                ActionFold,
		"dlourencobss: calls $0.03":       ActionCall,
		"KavarzE: checks":                 ActionCheck,
		"dlourencobss: bets $0.10":        ActionBet,
		"KavarzE: folds":                  ActionFold,
		"RE0309: folds":                   ActionFold,
		"pernadao1599: calls $0.10":       ActionCall,
		"dlourencobss: bets $0.27":        ActionBet,
		"pernadao1599: calls $0.27":       ActionCall,
		"dlourencobss: checks":            ActionCheck,
		"pernadao1599: checks":            ActionCheck,
	}

	for c, want := range cases {
		buffer := bytes.Buffer{}
		fmt.Fprintf(&buffer, "Post Scenario: %v", c)

		t.Run(buffer.String(), func(t *testing.T) {
			got, _ := actionTypeFromText([]byte(c))
			if got != want {
				t.Errorf("got %v, but wanted %v", got, want)
			}
		})
	}
}

func TestParseHandSummary(t *testing.T) {
	handText := handSummary

	summary, _ := parseHandSummary([]byte(handText))

	summaryWant := Summary{
		Pot:  0.36,
		Rake: 0.01,
		CommunityCards: [2]CommunityCards{
			{[3]Card{"Qc", "As", "3d"},
				Card("2h"),
				Card(""),
			},
			{}},
		Winners: []Winner{},
	}
	if !reflect.DeepEqual(summary, summaryWant) {
		t.Errorf("got %#v, but wanted %#v", summary, summaryWant)
	}
}

func TestParseMetadata(t *testing.T) {
	handText := testHands
	const rawETTime = "2025-01-21 15:51:32"
	loc, err := time.LoadLocation("America/New_York")

	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	handTime, _ := time.ParseInLocation(time.DateTime, rawETTime, loc)
	wantTime := handTime.UTC()

	metadata, err := parseMetaData([]byte(handText))

	if err != nil {
		t.Fatalf("parseMetaData returned error: %v", err)
	}

	metadataWant := Metadata{
		ID:         "254489598204",
		Date:       wantTime,
		ButtonSeat: 1,
	}

	if metadata != metadataWant {
		t.Errorf("\ngot %#v, but wanted %#v", metadata, metadataWant)
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
			got, err := actionPlayerNameFromText([]byte(c))

			if err != nil {
				t.Error("failed to parse name but should have")
			}

			if string(got) != want {
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
				got, _ := actionAmountFromText([]byte(c))
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
				got, _ := actionAmountFromText([]byte(c))
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
				_, err := actionAmountFromText([]byte(c))
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
Table 'Halley' 6-max Seat #1 is the button
Seat 1: test ($6000 in chips)
Seat 2: KavarzE ($3000 in chips)
Dealt to KavarzE [Ad Ac]
KavarzE: bets $2.33
*** SUMMARY ***
Total pot $0.25 | Rake $0.01
Seat 1: KavarzE won ($3.80)`)},
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
				[]Player{{
					Username:  "test",
					Cards:     [2]Card{"", ""},
					Seat:      1,
					ChipCount: 6000},

					{Username: "KavarzE",
						Cards:     [2]Card{"Ad", "Ac"},
						Seat:      2,
						ChipCount: 3000},
				},
				[]Action{
					{"KavarzE", 1, Preflop, ActionBet, 2.33},
				},
				Summary{
					[2]CommunityCards{}, 0, 0, []Winner{},
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
Table 'Euphemia II' 6-max (Play Money) Seat #3 is the button
Seat 1: test ($6000 in chips)
Seat 2: test2 ($3000 in chips)
Dealt to test [Ad Ac]
test: bets $2.33
*** SHOW DOWN ***
KavarzE collected $3.80 from pot
*** SUMMARY ***
Total pot $0.25 | Rake $0.01
Seat 1: KavarzE won ($3.80)`)

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
				Metadata{"123", time.Time{}.UTC(), 3},
				[]Player{
					{Username: "test", Cards: [2]Card{"Ad", "Ac"}, Seat: 1, ChipCount: 6000},
					{Username: "test2", Cards: [2]Card{"", ""}, Seat: 2, ChipCount: 3000}},
				[]Action{{"test", 1, Preflop, ActionBet, 2.33}},
				Summary{[2]CommunityCards{}, 0.25, 0.01, []Winner{{"KavarzE", 3.80, 1}}},
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

	got := handIDFromText([]byte(handData))
	want := []byte("6548679821301346841")

	if string(got) != string(want) {
		t.Errorf("got %v wanted %v", got, want)
	}
}

func TestActionsFromText(t *testing.T) {
	var dummyStreet = Flop
	order := 1

	handData := []byte(`Kavarz: bets $3`)

	got, _, err := parseActionLine(handData, &dummyStreet, &order)
	if err != nil {
		t.Error(err)
	}

	want := Action{
		PlayerName: "Kavarz",
		Order:      2,
		Street:     Flop,
		ActionType: ActionBet,
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

			got := dateTimeFromText([]byte(tt.test))
			if string(got) != tt.want {
				t.Errorf("got %v but wanted %v", got, tt.want)
			}

		})
	}
}

func TestParseDateTime(t *testing.T) {
	rawETTime := "2025-01-27 12:49:38"

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	got := parseDateTime([]byte(rawETTime))

	handTime, _ := time.ParseInLocation(time.DateTime, rawETTime, loc)
	want := handTime.UTC()

	if got != want {
		t.Errorf("\ngot %v wanted %v", got, want)
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
			got, err := seatIntFromText([]byte(tt.test))

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
	got := communityCardsFromText([]byte(handText), boardSignifier)

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
		{`Dealt to KavarzE [Js 5c]`, Player{"KavarzE", [2]Card{"Js", "5c"}, 0, 0}},
		{`Seat 6: KavarzE ($1.97 in chips) `, Player{"KavarzE", [2]Card{}, 6, 1.97}},
	}

	for _, tt := range cases {
		t.Run(tt.test, func(t *testing.T) {

			got, _, _ := parsePlayer([]byte(tt.test))

			if got != tt.want {
				t.Errorf("got %v but we wanted %v", got, tt.want)
			}
		})
	}

}

func TestPotFromText(t *testing.T) {

	t.Run("happy path pot", func(t *testing.T) {
		cases := []struct {
			test     string
			wantPot  float64
			wantRake float64
		}{
			{"Total pot $0.94 | Rake $0", 0.94, 0},
			{"Total pot $10.55 | Rake $0.94", 10.55, 0.94},
			{"Total pot $198.36 | Rake $10.22", 198.36, 10.22},
		}

		for _, tt := range cases {
			gotPot, gotRake, err := potFromText([]byte(tt.test))

			if err != nil {
				t.Error("expected nil error but got one")
			}

			if gotPot != tt.wantPot {
				t.Errorf("got %f wanted %f", gotPot, tt.wantPot)
			}

			if gotRake != tt.wantRake {
				t.Errorf("got %v, but wanted %v", gotPot, tt.wantRake)
			}
		}
	})

	t.Run("failing non-float strings", func(t *testing.T) {
		cases := []struct {
			test       string
			resultPot  float64
			resultRake float64
		}{
			{"Total pot oh no there's no float value here", 0, 0},
			{"Total pot $ unable to parsey", 0, 0},
			{"Total pot Rake $ unparsable", 0, 0},
		}

		for _, tt := range cases {
			gotPot, gotRake, err := potFromText([]byte(tt.test))

			if err == nil {
				t.Errorf("expected err but didn't get one. case: %v", tt.test)
			}

			if gotPot != tt.resultPot {
				t.Errorf("got %v, but wanted %v", gotPot, tt.resultPot)
			}

			if gotRake != tt.resultRake {
				t.Errorf("got %v, but wanted %v", gotPot, tt.resultRake)
			}
		}
	})

	t.Run("non total pot/rake line", func(t *testing.T) {
		gotPot, gotRake, err := potFromText([]byte("Seat 1: KavarzE won ($3.89)"))

		if gotPot != 0 {
			t.Errorf("wanted 0 but got %v", gotPot)
		}

		if gotRake != 0 {
			t.Errorf("wanted 0 but got %v", gotRake)
		}

		if err != nil {
			t.Errorf("wanted nil err but got %v", err)
		}

	})

}

func TestUpdateOrAppendPlayer(t *testing.T) {
	players := map[string]Player{
		"KavarzE": {"KavarzE", [2]Card{"", ""}, 1, 6.00},
		"Javormy": {"Javormy", [2]Card{"", ""}, 2, 33.00},
		"noob":    {"noob", [2]Card{"", ""}, 3, 4.00},
	}

	updateOrAddPlayer(
		players,
		Player{"KavarzE", [2]Card{"Ac", "Ad"}, 1, 6.00},
	)

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
	players := map[string]Player{
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

func TestNoShowdownWinner(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		street  Street
		want    []Winner
		wantErr bool
	}{
		{
			name:   "preflop fold winner, board 0",
			line:   "Seat 6: KavarzE collected ($0.12)",
			street: Preflop,
			want:   []Winner{{PlayerName: "KavarzE", Amount: 0.12, Board: 0}},
		},
		{
			name:   "flop fold winner, board 1",
			line:   "Seat 1: KavarzE (button) collected ($0.27)",
			street: Flop,
			want:   []Winner{{PlayerName: "KavarzE", Amount: 0.27, Board: 1}},
		},
		{
			name:   "turn fold winner, board 1",
			line:   "Seat 1: KavarzE (button) collected ($0.27)",
			street: Turn,
			want:   []Winner{{PlayerName: "KavarzE", Amount: 0.27, Board: 1}},
		},
		{
			name:   "river fold winner, board 1",
			line:   "Seat 1: KavarzE (button) collected ($0.27)",
			street: River,
			want:   []Winner{{PlayerName: "KavarzE", Amount: 0.27, Board: 1}},
		},
		{
			name:   "non-matching line returns empty",
			line:   "Seat 2: SpieWNogach (small blind) folded before Flop",
			street: Preflop,
			want:   []Winner{},
		},
		{
			name:    "invalid currency returns error",
			line:    "Seat 6: KavarzE collected (£0.12)",
			street:  Preflop,
			want:    []Winner{},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := noShowdownWinner([]byte(tt.line), tt.street)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWinnerFromLine(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		boardNum int
		want     []Winner
		wantErr  bool
	}{
		{
			name:     "rio winner board 1",
			line:     "Jero0987 collected $5.12 from pot",
			boardNum: 1,
			want:     []Winner{{PlayerName: "Jero0987", Amount: 5.12, Board: 1}},
		},
		{
			name:     "rit second board winner",
			line:     "ribo7falani collected $5.12 from pot",
			boardNum: 2,
			want:     []Winner{{PlayerName: "ribo7falani", Amount: 5.12, Board: 2}},
		},
		{
			name:     "non-matching line returns empty",
			line:     "ribo7falani: shows [Kh Jd] (a full house, Jacks full of Kings)",
			boardNum: 1,
			want:     []Winner{},
		},
		{
			name:     "summary collected line not matched",
			line:     "Seat 1: KavarzE (button) collected ($0.27)",
			boardNum: 1,
			want:     []Winner{},
		},
		{
			name:     "invalid currency returns error",
			line:     "Jero0987 collected £5.12 from pot",
			boardNum: 1,
			want:     []Winner{},
			wantErr:  true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := winnerFromLine([]byte(tt.line), tt.boardNum)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExtractWinners(t *testing.T) {

	cases := []struct {
		name    string
		line    string
		state   ShowdownState
		street  Street
		want    []Winner
		wantErr bool
	}{
		{
			name:   "no showdown, preflop fold winner",
			line:   "Seat 6: KavarzE collected ($0.12)",
			state:  noShowdown,
			street: Preflop,
			want:   []Winner{{PlayerName: "KavarzE", Amount: 0.12, Board: 0}},
		},
		{
			name:   "no showdown, postflop winner board 1",
			line:   "Seat 1: KavarzE (button) collected ($0.27)",
			state:  noShowdown,
			street: Turn,
			want:   []Winner{{PlayerName: "KavarzE", Amount: 0.27, Board: 1}},
		},
		{
			name:   "no showdown, non-winner summary line ignored",
			line:   "Seat 2: SpieWNogach (small blind) folded before Flop",
			state:  noShowdown,
			street: Preflop,
			want:   []Winner{},
		},
		{
			name:   "rio winner",
			line:   "Jero0987 collected $5.12 from pot",
			state:  rio,
			street: River,
			want:   []Winner{{PlayerName: "Jero0987", Amount: 5.12, Board: 1}},
		},
		{
			name:   "rit first board winner",
			line:   "Jero0987 collected $5.12 from pot",
			state:  ritFirstBoard,
			street: River,
			want:   []Winner{{PlayerName: "Jero0987", Amount: 5.12, Board: 1}},
		},
		{
			name:   "rit second board winner",
			line:   "ribo7falani collected $5.12 from pot",
			state:  ritSecondBoard,
			street: River,
			want:   []Winner{{PlayerName: "ribo7falani", Amount: 5.12, Board: 2}},
		},
		{
			name:   "non-winner line in showdown ignored",
			line:   "ribo7falani: shows [Kh Jd] (a full house, Jacks full of Kings)",
			state:  rio,
			street: River,
			want:   []Winner{},
		},
		{
			name:    "invalid amount returns error",
			line:    "Seat 1: KavarzE (button) collected (£0.27)",
			state:   noShowdown,
			street:  Turn,
			want:    []Winner{},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractWinners([]byte(tt.line), tt.state, tt.street)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExtractButtonSeatFromText(t *testing.T) {
	cases := []struct {
		test string
		want int64
	}{
		{`PokerStars Hand #254446123323:  Hold'em No Limit ($0.02/$0.05 USD) - 2025/01/19 12:38:55 WET [2025/01/19 7:38:55 ET]
Table 'Wei III' 6-max Seat #1 is the button
Seat 1: maximoIV ($5.20 in chips)
Seat 2: dlourencobss ($4.94 in chips)
Seat 3: KavarzE ($5 in chips)
Seat 4: arsad725 ($5.49 in chips)
`, 1},
		{`PokerStars Hand #231244441:  Hold'em No Limit ($0.02/$0.05 USD) - 2025/01/19 12:38:55 WET [2025/01/19 7:38:55 ET]
Table 'Wei III' 6-max Seat #9 is the button`, 9},
		{`6-max Seat #5 is the button`, 5},
	}

	for _, tt := range cases {
		got, err := extractButtonSeatFromText([]byte(tt.test))

		if err != nil {
			t.Errorf("wanted nil error but got %v", err)
		}

		if got != tt.want {
			t.Errorf("wanted %v but got %v", tt.want, got)
		}
	}
}

func TestSubstringBetween(t *testing.T) {
	cases := []struct {
		test  string
		start string
		end   string
		want  string
	}{
		{
			test:  "PokerStars Zoom Hand #254489598204:  Hold'em No Limit ($0.02/$0.05) - 2025/01/21 20:51:32 WET [2025/01/21 15:51:32 ET]",
			start: " [",
			end:   " ET]",
			want:  "2025/01/21 15:51:32",
		},
		{
			test:  "Dealt to KavarzE [Ac Dc]",
			start: " [",
			end:   "]",
			want:  "Ac Dc",
		},
		{
			test: `PokerStars Zoom Hand #257507385322:  Hold'em No Limit ($0.01/$0.02) - 2025/08/27 18:30:17 WET [2025/08/27 13:30:17 ET]
Table 'Halley' 6-max Seat #1 is the button
Seat 1: TSCardinals ($2.02 in chips) 
`,
			start: "Seat #",
			end:   " is the button",
			want:  "1",
		},
	}

	for _, tt := range cases {
		got := substringBetween([]byte(tt.test), []byte(tt.start), []byte(tt.end))

		if string(got) != tt.want {
			t.Errorf("wanted %v but got %v", tt.want, got)
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

const handSummary string = `Total pot $0.36 | Rake $0.01
Board [Qc As 3d 2h]
Seat 1: JDfq28 (button) folded on the Turn
Seat 2: kv_def (small blind) collected ($0.35)
Seat 3: KavarzE (big blind) folded on the Turn
Seat 4: MGPN folded before Flop (didn't bet)
Seat 5: ikin23 folded before Flop (didn't bet)
Seat 6: honda589 folded before Flop (didn't bet)`

const uncalledBetHand string = `PokerStars Zoom Hand #257507385322:  Hold'em No Limit ($0.01/$0.02) - 2025/08/27 18:30:17 WET [2025/08/27 13:30:17 ET]
Table 'Halley' 6-max Seat #1 is the button
Seat 1: TSCardinals ($2.02 in chips) 
Seat 2: Jimmey54 ($2.21 in chips) 
Seat 3: nm8800 ($2.31 in chips) 
Seat 4: Chewbacca97 ($1.08 in chips) 
Seat 5: KavarzE ($2.08 in chips) 
Seat 6: haeorm ($6.26 in chips) 
Jimmey54: posts small blind $0.01
nm8800: posts big blind $0.02
*** HOLE CARDS ***
Dealt to KavarzE [8s As]
Chewbacca97: folds 
KavarzE: raises $0.04 to $0.06
haeorm: folds 
TSCardinals: calls $0.06
Jimmey54: folds 
nm8800: folds 
*** FLOP *** [Tc 4h 6h]
KavarzE: bets $0.04
TSCardinals: calls $0.04
*** TURN *** [Tc 4h 6h] [5c]
KavarzE: checks 
TSCardinals: bets $0.17
KavarzE: folds 
Uncalled bet ($0.17) returned to TSCardinals
TSCardinals collected $0.22 from pot
TSCardinals: doesn't show hand 
*** SUMMARY ***
Total pot $0.23 | Rake $0.01 
Board [Tc 4h 6h 5c]
Seat 1: TSCardinals (button) collected ($0.22)
Seat 2: Jimmey54 (small blind) folded before Flop
Seat 3: nm8800 (big blind) folded before Flop
Seat 4: Chewbacca97 folded before Flop (didn't bet)
Seat 5: KavarzE folded on the Turn
Seat 6: haeorm folded before Flop (didn't bet)`

const runItTwice string = `PokerStars Zoom Hand #254607988518:  Hold'em No Limit ($0.02/$0.05) - 2025/01/29 16:30:35 WET [2025/01/29 11:30:35 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: TurivVB240492 ($1.94 in chips)
Seat 2: KavarzE ($15.14 in chips)
Seat 3: RoMike2 ($5.07 in chips)
Seat 4: hiroakin ($5 in chips)
Seat 5: ThxWasOby3 ($5.22 in chips)
Seat 6: VLSALT ($5 in chips)
KavarzE: posts small blind $0.02
RoMike2: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [Jc Js]
hiroakin: folds
ThxWasOby3: raises $0.10 to $0.15
VLSALT: folds
TurivVB240492: folds
KavarzE: raises $0.45 to $0.60
RoMike2: folds
ThxWasOby3: raises $0.72 to $1.32
KavarzE: calls $0.72
*** FLOP *** [7d 2h 8h]
KavarzE: checks
ThxWasOby3: checks
*** TURN *** [7d 2h 8h] [Jh]
KavarzE: bets $1.81
ThxWasOby3: raises $2.09 to $3.90 and is all-in
KavarzE: calls $2.09
*** FIRST RIVER *** [7d 2h 8h Jh] [3d]
*** SECOND RIVER *** [7d 2h 8h Jh] [Qh]
*** FIRST SHOW DOWN ***
KavarzE: shows [Jc Js] (three of a kind, Jacks)
ThxWasOby3: shows [Ah Qd] (high card Ace)
KavarzE collected $5.03 from pot
*** SECOND SHOW DOWN ***
KavarzE: shows [Jc Js] (three of a kind, Jacks)
ThxWasOby3: shows [Ah Qd] (a flush, Ace high)
ThxWasOby3 collected $5.02 from pot
*** SUMMARY ***
Total pot $10.49 | Rake $0.44
Hand was run twice
FIRST Board [7d 2h 8h Jh 3d]
SECOND Board [7d 2h 8h Jh Qh]
Seat 1: TurivVB240492 (button) folded before Flop (didn't bet)
Seat 2: KavarzE (small blind) showed [Jc Js] and won ($5.03) with three of a kind, Jacks, and lost with three of a kind, Jacks
Seat 3: RoMike2 (big blind) folded before Flop
Seat 4: hiroakin folded before Flop (didn't bet)
Seat 5: ThxWasOby3 showed [Ah Qd] and lost with high card Ace, and won ($5.02) with a flush, Ace high
Seat 6: VLSALT folded before Flop (didn't bet)`

const runItTwicePlayerWonBothBoards string = `PokerStars Zoom Hand #254449744546:  Hold'em No Limit ($0.02/$0.05) - 2025/01/19 16:36:38 WET [2025/01/19 11:36:38 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: AsmAngAmAngo ($6.95 in chips) 
Seat 2: loto_insane ($5 in chips) 
Seat 3: KavarzE ($7.11 in chips) 
Seat 4: Braghinn ($5.72 in chips) 
Seat 5: R.S.P747 ($5.51 in chips) 
Seat 6: Gatzin ($6.88 in chips) 
loto_insane: posts small blind $0.02
KavarzE: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [As Jc]
Braghinn: raises $0.06 to $0.11
R.S.P747: folds 
Gatzin: calls $0.11
AsmAngAmAngo: folds 
loto_insane: calls $0.09
KavarzE: raises $0.89 to $1
Braghinn: folds 
Gatzin: calls $0.89
loto_insane: folds 
*** FLOP *** [Js 7s 8c]
KavarzE: bets $1.60
Gatzin: calls $1.60
*** TURN *** [Js 7s 8c] [6h]
KavarzE: bets $4.51 and is all-in
Gatzin: calls $4.28 and is all-in
Uncalled bet ($0.23) returned to KavarzE
*** FIRST RIVER *** [Js 7s 8c 6h] [6d]
*** SECOND RIVER *** [Js 7s 8c 6h] [Ks]
*** FIRST SHOW DOWN ***
KavarzE: shows [As Jc] (two pair, Jacks and Sixes)
Gatzin: shows [Qh Jh] (two pair, Jacks and Sixes - lower kicker)
KavarzE collected $6.70 from pot
*** SECOND SHOW DOWN ***
KavarzE: shows [As Jc] (a pair of Jacks)
Gatzin: shows [Qh Jh] (a pair of Jacks - lower kicker)
KavarzE collected $6.70 from pot
*** SUMMARY ***
Total pot $13.98 | Rake $0.58 
Hand was run twice
FIRST Board [Js 7s 8c 6h 6d]
SECOND Board [Js 7s 8c 6h Ks]
Seat 1: AsmAngAmAngo (button) folded before Flop (didn't bet)
Seat 2: loto_insane (small blind) folded before Flop
Seat 3: KavarzE (big blind) showed [As Jc] and won ($6.70) with two pair, Jacks and Sixes, and won ($6.70) with a pair of Jacks
Seat 4: Braghinn folded before Flop
Seat 5: R.S.P747 folded before Flop (didn't bet)
Seat 6: Gatzin showed [Qh Jh] and lost with two pair, Jacks and Sixes, and lost with a pair of Jacks`

const allFoldedBeforeFlop string = `PokerStars Zoom Hand #254626485418:  Hold'em No Limit ($0.02/$0.05) - 2025/01/30 19:50:17 WET [2025/01/30 14:50:17 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: OoJohnStevensoO ($6.24 in chips) 
Seat 2: bk4crs ($9.22 in chips) 
Seat 3: KavarzE ($5 in chips) 
Seat 4: FabuTK ($4.35 in chips) 
Seat 5: getaddicted ($6.59 in chips) 
Seat 6: ilbeback2017 ($19.69 in chips) 
bk4crs: posts small blind $0.02
KavarzE: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [Qd 5c]
FabuTK: folds 
getaddicted: folds 
ilbeback2017: folds 
OoJohnStevensoO: raises $0.06 to $0.11
bk4crs: folds 
KavarzE: folds 
Uncalled bet ($0.06) returned to OoJohnStevensoO
OoJohnStevensoO collected $0.12 from pot
OoJohnStevensoO: doesn't show hand 
*** SUMMARY ***
Total pot $0.12 | Rake $0 
Seat 1: OoJohnStevensoO (button) collected ($0.12)
Seat 2: bk4crs (small blind) folded before Flop
Seat 3: KavarzE (big blind) folded before Flop
Seat 4: FabuTK folded before Flop (didn't bet)
Seat 5: getaddicted folded before Flop (didn't bet)
Seat 6: ilbeback2017 folded before Flop (didn't bet)`

const ritEdgeCaseHand string = `PokerStars Zoom Hand #254626500457:  Hold'em No Limit ($0.02/$0.05) - 2025/01/30 19:51:09 WET [2025/01/30 14:51:09 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: Zutuzutu_90 ($7.31 in chips) 
Seat 2: KavarzE ($5 in chips) 
Seat 3: darchas ($5 in chips) 
Seat 4: soyjuliansito ($5.03 in chips) 
Seat 5: SpieWNogach ($5.07 in chips) 
Seat 6: Trogloditapubg ($4.75 in chips) 
KavarzE: posts small blind $0.02
darchas: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [9s Ks]
soyjuliansito: folds 
SpieWNogach: folds 
Trogloditapubg: folds 
Zutuzutu_90: raises $0.07 to $0.12
KavarzE: raises $0.33 to $0.45
darchas: folds 
Zutuzutu_90: calls $0.33
*** FLOP *** [Ts 2d 8s]
KavarzE: bets $0.30
Zutuzutu_90: raises $0.45 to $0.75
KavarzE: calls $0.45
*** TURN *** [Ts 2d 8s] [7h]
KavarzE: bets $3.80 and is all-in
Zutuzutu_90: calls $3.80
*** FIRST RIVER *** [Ts 2d 8s 7h] [Kh]
*** SECOND RIVER *** [Ts 2d 8s 7h] [6d]
*** FIRST SHOW DOWN ***
KavarzE: shows [9s Ks] (a pair of Kings)
Zutuzutu_90: shows [Tc 9c] (a pair of Tens)
KavarzE collected $4.82 from pot
*** SECOND SHOW DOWN ***
KavarzE: shows [9s Ks] (a straight, Six to Ten)
Zutuzutu_90: shows [Tc 9c] (a straight, Six to Ten)
KavarzE collected $2.41 from pot
Zutuzutu_90 collected $2.37 from pot
*** SUMMARY ***
Total pot $10.05 | Rake $0.45 
Hand was run twice
FIRST Board [Ts 2d 8s 7h Kh]
SECOND Board [Ts 2d 8s 7h 6d]
Seat 1: Zutuzutu_90 (button) showed [Tc 9c] and lost with a pair of Tens, and won ($2.37) with a straight, Six to Ten
Seat 2: KavarzE (small blind) showed [9s Ks] and won ($4.82) with a pair of Kings, and won ($2.41) with a straight, Six to Ten
Seat 3: darchas (big blind) folded before Flop
Seat 4: soyjuliansito folded before Flop (didn't bet)
Seat 5: SpieWNogach folded before Flop (didn't bet)
Seat 6: Trogloditapubg folded before Flop (didn't bet)`
