package pokerhud_test

import (
	"errors"
	"io/fs"
	"pokerhud"
	"reflect"
	"testing"
	"testing/fstest"
	"time"
)

func TestHandHistoriesFromFS(t *testing.T) {
	t.Run("multiple files and multiple hands returns correct number of hands", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"zoom.txt":      {Data: []byte(zoomHand1)},
			"cash game.txt": {Data: []byte(cashGame1)},
		}

		handHistory, err := pokerhud.HandHistoryFromFS(fileSystem)

		if len(handHistory) != 4 {
			t.Errorf("wanted 4 hands, got %d", len(handHistory))
		}

		if err != nil {
			t.Error("expected no errors but got one")
		}
	})

	// t.Run("after successfully parsing the file, the HH file is moved to assigned dir", func(t *testing.T) {
	// 	fileSystem := fstest.MapFS{
	// 		"zoom.txt":      {Data: []byte(zoomHand1)},
	// 		"cash game.txt": {Data: []byte(cashGame1)},
	// 		"new-folder/" : &fstest.MapFile{Mode: fs.ModeDir},
	// 	}

	// 	pokerhud.HandHistoryFromFS(fileSystem)

	// 	want := fstest.MapFS{
	// 		"new-folder/" : &fstest.MapFile{
	// 			Mode: fs.ModeDir,
	// 		},
	// 	}

	// 	want["new-folder/zoom.txt"] = &fstest.MapFile{
	// 		Data: []byte(zoomHand1),
	// 		Mode: 0664,
	// 	}

	// 	want["new-folder/cash game.txt"] = &fstest.MapFile{
	// 		Data: []byte(cashGame1),
	// 		Mode: 0664,
	// 	}
		

	// 	if reflect.DeepEqual(fileSystem, want) {
	// 		t.Errorf("got %#v, but wanted %#v", fileSystem, want)
	// 	}

	// })

	t.Run("failing filesystem", func(t *testing.T) {
		fileSystem := failingFS{}

		_, err := pokerhud.HandHistoryFromFS(fileSystem)

		if err == nil {
			t.Fatal("expected an err but didn't get one")
		}
	})

	t.Run("test when one file fails", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"zoom.txt":      {Data: []byte(zoomHand1)},
			"cash game.txt": {Data: []byte(cashGame1)},
			"failure.txt":   {Data: []byte("not a hand")},
		}

		handHistory, _ := pokerhud.HandHistoryFromFS(fileSystem)

		if len(handHistory) != 4 {
			t.Errorf("wanted 2 hands, got %d", len(handHistory))
		}

	})

	t.Run("hand data is correctly parsed from a text file", func(t *testing.T) {
		fileSystem := fstest.MapFS{
			"Wei III": {Data: []byte(cashGame2)},
		}

		handHistory, _ := pokerhud.HandHistoryFromFS(fileSystem)
		handTime, _ := time.Parse(time.DateTime, "2025-01-19 12:38:55")

		got := handHistory[0]
		want := pokerhud.Hand{
			ID:      "254446123323",
			Date:    handTime.Local(),
			Players: []pokerhud.Player{{"maximoIV"}, {"dlourencobss"}, {"KavarzE"}, {"arsad725"}, {"RE0309"}, {"pernadao1599"}},
			Actions: []pokerhud.Action{
				actionBuildHelper("dlourencobss", pokerhud.Posts, pokerhud.Preflop, 1, 0.02),
				actionBuildHelper("KavarzE", pokerhud.Posts, pokerhud.Preflop, 2, 0.05),
				actionBuildHelper("arsad725", pokerhud.Folds, pokerhud.Preflop, 3, 0),
				actionBuildHelper("RE0309", pokerhud.Calls, pokerhud.Preflop, 4, 0.05),
				actionBuildHelper("pernadao1599", pokerhud.Calls, pokerhud.Preflop, 5, 0.05),
				actionBuildHelper("maximoIV", pokerhud.Folds, pokerhud.Preflop, 6, 0),
				actionBuildHelper("dlourencobss", pokerhud.Calls, pokerhud.Preflop, 7, 0.03),
				actionBuildHelper("KavarzE", pokerhud.Checks, pokerhud.Preflop, 8, 0),
				actionBuildHelper("dlourencobss", pokerhud.Bets, pokerhud.Flop, 9, 0.10),
				actionBuildHelper("KavarzE", pokerhud.Folds, pokerhud.Flop, 10, 0),
				actionBuildHelper("RE0309", pokerhud.Folds, pokerhud.Flop, 11, 0),
				actionBuildHelper("pernadao1599", pokerhud.Calls, pokerhud.Flop, 12, 0.10),
				actionBuildHelper("dlourencobss", pokerhud.Bets, pokerhud.Turn, 13, 0.27),
				actionBuildHelper("pernadao1599", pokerhud.Calls, pokerhud.Turn, 14, 0.27),
				actionBuildHelper("dlourencobss", pokerhud.Checks, pokerhud.River, 15, 0),
				actionBuildHelper("pernadao1599", pokerhud.Checks, pokerhud.River, 16, 0),
			},
			HeroCards: "2s 5d",
		}

		assertHand(t, got, want)
	})
}

func assertHand(t *testing.T, got, want pokerhud.Hand) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %#v, \n\ngot %#v", want, got)
	}
}

func BenchmarkHandHistoryFromFS(b *testing.B) {
	pokerhud.HandHistoryFromFS(fstest.MapFS{
		"zoom.txt":      {Data: []byte(zoomHand1)},
		"cash game.txt": {Data: []byte(cashGame1)},
		"failure.txt":   {Data: []byte("not a hand")},
	})
}

func actionBuildHelper(playerName string, actionType pokerhud.ActionType, street pokerhud.Street, order int, amount float64) pokerhud.Action {
	return pokerhud.Action{
		Player: pokerhud.Player{
			Username: playerName,
		},
		ActionType: actionType,
		Street:     street,
		Order:      order,
		Amount:     amount,
	}
}

type failingFS struct{}

func (f failingFS) Open(string) (fs.File, error) {
	return nil, errors.New("oh no i always fail")
}

type failingFile struct{}

func (f failingFile) Open(string) {

}

const zoomHand1 string = `PokerStars Zoom Hand #254445778475:  Hold'em No Limit ($0.02/$0.05) - 2025/01/19 12:00:43 WET [2025/01/19 7:00:43 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: tiodor1972 ($7 in chips)
Seat 2: Siku44 ($4.32 in chips)
Seat 3: BorisBorisPaul ($7.20 in chips)
Seat 4: KavarzE ($5 in chips)
Seat 5: JacDs ($6.51 in chips)
Seat 6: andots7 ($5.07 in chips)
Siku44: posts small blind $0.02
BorisBorisPaul: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [5d 4d]
KavarzE: raises $0.08 to $0.13
JacDs: folds
andots7: folds
tiodor1972: folds
Siku44: folds
BorisBorisPaul: calls $0.08
*** FLOP *** [Jc 6d 6h]
BorisBorisPaul: checks
KavarzE: bets $0.09
BorisBorisPaul: raises $0.11 to $0.20
KavarzE: calls $0.11
*** TURN *** [Jc 6d 6h] [8d]
BorisBorisPaul: bets $0.45
KavarzE: calls $0.45
*** RIVER *** [Jc 6d 6h 8d] [2c]
BorisBorisPaul: bets $0.95
KavarzE: folds
Uncalled bet ($0.95) returned to BorisBorisPaul
BorisBorisPaul collected $1.51 from pot
BorisBorisPaul: doesn't show hand
*** SUMMARY ***
Total pot $1.58 | Rake $0.07
Board [Jc 6d 6h 8d 2c]
Seat 1: tiodor1972 (button) folded before Flop (didn't bet)
Seat 2: Siku44 (small blind) folded before Flop
Seat 3: BorisBorisPaul (big blind) collected ($1.51)
Seat 4: KavarzE folded on the River
Seat 5: JacDs folded before Flop (didn't bet)
Seat 6: andots7 folded before Flop (didn't bet)



PokerStars Zoom Hand #254445789939:  Hold'em No Limit ($0.02/$0.05) - 2025/01/19 12:02:03 WET [2025/01/19 7:02:03 ET]
Table 'Donati' 6-max Seat #1 is the button
Seat 1: KavarzE ($5 in chips)
Seat 2: parktjdwn ($6.90 in chips)
Seat 3: tomato_yorumy ($5.27 in chips)
Seat 4: cotyara1986 ($5.66 in chips)
Seat 5: julesAAAA ($7.19 in chips)
Seat 6: Yatoro_7 ($5.26 in chips)
parktjdwn: posts small blind $0.02
tomato_yorumy: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [Jc Tc]
cotyara1986: folds
julesAAAA: folds
Yatoro_7: folds
KavarzE: raises $0.08 to $0.13
parktjdwn: raises $0.40 to $0.53
tomato_yorumy: folds
KavarzE: calls $0.40
*** FLOP *** [Js 5h Kh]
parktjdwn: bets $0.35
KavarzE: calls $0.35
*** TURN *** [Js 5h Kh] [9d]
parktjdwn: checks
KavarzE: checks
*** RIVER *** [Js 5h Kh 9d] [2d]
parktjdwn: bets $1.30
KavarzE: calls $1.30
*** SHOW DOWN ***
parktjdwn: shows [5s 6s] (a pair of Fives)
KavarzE: shows [Jc Tc] (a pair of Jacks)
KavarzE collected $4.23 from pot
*** SUMMARY ***
Total pot $4.41 | Rake $0.18
Board [Js 5h Kh 9d 2d]
Seat 1: KavarzE (button) showed [Jc Tc] and won ($4.23) with a pair of Jacks
Seat 2: parktjdwn (small blind) showed [5s 6s] and lost with a pair of Fives
Seat 3: tomato_yorumy (big blind) folded before Flop
Seat 4: cotyara1986 folded before Flop (didn't bet)
Seat 5: julesAAAA folded before Flop (didn't bet)
Seat 6: Yatoro_7 folded before Flop (didn't bet)`

const cashGame1 string = `PokerStars Hand #254446123323:  Hold'em No Limit ($0.02/$0.05 USD) - 2025/01/19 12:38:55 WET [2025/01/19 7:38:55 ET]
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
Seat 6: pernadao1599 showed [Jh Qc] and won ($0.89) with a pair of Jacks



PokerStars Hand #254446129322:  Hold'em No Limit ($0.02/$0.05 USD) - 2025/01/19 12:39:33 WET [2025/01/19 7:39:33 ET]
Table 'Wei III' 6-max Seat #2 is the button
Seat 4: arsad725 ($5.49 in chips) is sitting out
Seat 1: maximoIV ($5.20 in chips)
Seat 2: dlourencobss ($4.52 in chips)
Seat 3: KavarzE ($5 in chips)
Seat 5: RE0309 ($4.58 in chips)
Seat 6: pernadao1599 ($3.90 in chips)
KavarzE: posts small blind $0.02
RE0309: posts big blind $0.05
*** HOLE CARDS ***
Dealt to KavarzE [9c Kh]
pernadao1599: folds
maximoIV: raises $0.07 to $0.12
dlourencobss: folds
KavarzE: calls $0.10
RE0309: folds
*** FLOP *** [8h 6s 5d]
KavarzE: checks
dlourencobss leaves the table
maximoIV: checks
*** TURN *** [8h 6s 5d] [Jh]
bananen333 joins the table at seat #2
KavarzE: bets $0.09
maximoIV: calls $0.09
*** RIVER *** [8h 6s 5d Jh] [Ks]
KavarzE: bets $0.18
maximoIV: folds
Uncalled bet ($0.18) returned to KavarzE
KavarzE collected $0.45 from pot
*** SUMMARY ***
Total pot $0.47 | Rake $0.02
Board [8h 6s 5d Jh Ks]
Seat 4: arsad725
Seat 1: maximoIV folded on the River
Seat 2: dlourencobss (button) folded before Flop (didn't bet)
Seat 3: KavarzE (small blind) collected ($0.45)
Seat 5: RE0309 (big blind) folded before Flop
Seat 6: pernadao1599 folded before Flop (didn't bet)`

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