package pokerhud

import (
	"fmt"
)

type HandDataReader interface {
	Read()
}

type HandData struct {
	text string
}

// func TestParseHandData(t *testing.T) {
// 	testHandData := HandData{text: `PokerStars Zoom Hand #254445778475:  Hold'em No Limit ($0.02/$0.05) - 2025/01/19 12:00:43 WET [2025/01/19 7:00:43 ET]
// Table 'Donati' 6-max Seat #1 is the button
// Seat 1: tiodor1972 ($7 in chips)
// Seat 2: Siku44 ($4.32 in chips)
// Seat 3: BorisBorisPaul ($7.20 in chips)
// Seat 4: KavarzE ($5 in chips)
// Seat 5: JacDs ($6.51 in chips)
// Seat 6: andots7 ($5.07 in chips)
// Siku44: posts small blind $0.02
// BorisBorisPaul: posts big blind $0.05
// *** HOLE CARDS ***
// Dealt to KavarzE [5d 4d]
// KavarzE: raises $0.08 to $0.13
// JacDs: folds
// andots7: folds
// tiodor1972: folds
// Siku44: folds
// BorisBorisPaul: calls $0.08
// *** FLOP *** [Jc 6d 6h]
// BorisBorisPaul: checks
// KavarzE: bets $0.09
// BorisBorisPaul: raises $0.11 to $0.20
// KavarzE: calls $0.11
// *** TURN *** [Jc 6d 6h] [8d]
// BorisBorisPaul: bets $0.45
// KavarzE: calls $0.45
// *** RIVER *** [Jc 6d 6h 8d] [2c]
// BorisBorisPaul: bets $0.95
// KavarzE: folds
// Uncalled bet ($0.95) returned to BorisBorisPaul
// BorisBorisPaul collected $1.51 from pot
// BorisBorisPaul: doesn't show hand
// *** SUMMARY ***
// Total pot $1.58 | Rake $0.07
// Board [Jc 6d 6h 8d 2c]
// Seat 1: tiodor1972 (button) folded before Flop (didn't bet)
// Seat 2: Siku44 (small blind) folded before Flop
// Seat 3: BorisBorisPaul (big blind) collected ($1.51)
// Seat 4: KavarzE folded on the River
// Seat 5: JacDs folded before Flop (didn't bet)
// Seat 6: andots7 folded before Flop (didn't bet)

// PokerStars Zoom Hand #254445789939:  Hold'em No Limit ($0.02/$0.05) - 2025/01/19 12:02:03 WET [2025/01/19 7:02:03 ET]
// Table 'Donati' 6-max Seat #1 is the button
// Seat 1: KavarzE ($5 in chips)
// Seat 2: parktjdwn ($6.90 in chips)
// Seat 3: tomato_yorumy ($5.27 in chips)
// Seat 4: cotyara1986 ($5.66 in chips)
// Seat 5: julesAAAA ($7.19 in chips)
// Seat 6: Yatoro_7 ($5.26 in chips)
// parktjdwn: posts small blind $0.02
// tomato_yorumy: posts big blind $0.05
// *** HOLE CARDS ***
// Dealt to KavarzE [Jc Tc]
// cotyara1986: folds
// julesAAAA: folds
// Yatoro_7: folds
// KavarzE: raises $0.08 to $0.13
// parktjdwn: raises $0.40 to $0.53
// tomato_yorumy: folds
// KavarzE: calls $0.40
// *** FLOP *** [Js 5h Kh]
// parktjdwn: bets $0.35
// KavarzE: calls $0.35
// *** TURN *** [Js 5h Kh] [9d]
// parktjdwn: checks
// KavarzE: checks
// *** RIVER *** [Js 5h Kh 9d] [2d]
// parktjdwn: bets $1.30
// KavarzE: calls $1.30
// *** SHOW DOWN ***
// parktjdwn: shows [5s 6s] (a pair of Fives)
// KavarzE: shows [Jc Tc] (a pair of Jacks)
// KavarzE collected $4.23 from pot
// *** SUMMARY ***
// Total pot $4.41 | Rake $0.18
// Board [Js 5h Kh 9d 2d]
// Seat 1: KavarzE (button) showed [Jc Tc] and won ($4.23) with a pair of Jacks
// Seat 2: parktjdwn (small blind) showed [5s 6s] and lost with a pair of Fives
// Seat 3: tomato_yorumy (big blind) folded before Flop
// Seat 4: cotyara1986 folded before Flop (didn't bet)
// Seat 5: julesAAAA folded before Flop (didn't bet)
// Seat 6: Yatoro_7 folded before Flop (didn't bet)
// `}

// }

func (h HandData) Read() (int, error) {
	fmt.Println(h.text)
	return 1, nil
}
