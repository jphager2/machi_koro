package main

import (
	"fmt"
)

type effect struct {
	Priority    int
	Description func() string
	Call        func(card supplyCard, rlr *player, p *player, c int, specialRoll int)
}

type prereq struct {
	Desc string
	Call func(card supplyCard, rlr *player, p *player, c int, specialRoll int) bool
}

var nullPrereq = prereq{
	Desc: "",

	Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) bool {
		return true
	},
}

func landmarkCardAgumentedPayout(payout int, card supplyCard, p *player) int {
	if p.LandmarkCards["Shopping Mall"] && (card.Icon == "Cup" || card.Icon == "Bread") {
		return payout + 1
	}
	return payout
}

func newBankPayoutWithPrereq(payout int, onlyCurrent bool, fromBank bool, pr prereq) effect {
	var coins string
	var receiverName string

	if payout == 1 {
		coins = "1 coin"
	} else {
		coins = fmt.Sprintf("%d coins", payout)
	}

	if onlyCurrent {
		receiverName = "your turn only"
	} else {
		receiverName = "anyone's turn"
	}

	return effect{
		Priority: 1,

		Description: func() string {
			if fromBank {
				return pr.Desc + fmt.Sprintf("Get %s from the bank on %s", coins, receiverName)
			}

			return pr.Desc + fmt.Sprintf("Pay %s to the bank on %s", coins, receiverName)
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			if onlyCurrent && p != rlr {
				return
			}
			if !pr.Call(card, rlr, p, c, specialRoll) {
				return
			}

			totalPayout := landmarkCardAgumentedPayout(payout, card, p) * c

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}
}

func newAllBankPayout(r int) effect {
	return newBankPayoutWithPrereq(r, false, true, nullPrereq)
}

func newAllBankPayoutWithPrereq(r int, pr prereq) effect {
	return newBankPayoutWithPrereq(r, false, true, pr)
}

func newRollerBankPayout(r int) effect {
	return newBankPayoutWithPrereq(r, true, true, nullPrereq)
}

func newBankRollerPayout(r int) effect {
	return newBankPayoutWithPrereq(r, true, false, nullPrereq)
}

func newRollerBankPayoutWithPrereq(r int, pr prereq) effect {
	return newBankPayoutWithPrereq(r, true, true, pr)
}

func newRollerPayoutWithPrereq(payout int, pr prereq) effect {
	var coins string

	if payout == 1 {
		coins = "1 coin"
	} else {
		coins = fmt.Sprintf("%d coins", payout)
	}

	return effect{
		Priority: 0,

		Description: func() string {
			desc := pr.Desc + fmt.Sprintf("Get %s from the player who rolled the dice", coins)
			return desc
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			if p == rlr {
				return
			}
			if !pr.Call(card, rlr, p, c, specialRoll) {
				return
			}

			totalPayout := landmarkCardAgumentedPayout(payout, card, p) * c

			fmt.Printf("Player %d gets %d coins from the player %d [%s].\n", p.ID, totalPayout, rlr.ID, card.Name)
			remainder := rlr.Coins.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Roller did not have enough money. Missing: %d\n", remainder)
			}
		},
	}
}

func newLandmarkPrereq(name string, forRoller bool) prereq {
	return prereq{
		Desc: fmt.Sprintf("If you have the [%s] landmark. ", name),
		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) bool {
			if forRoller {
				return rlr.LandmarkCards[name]
			}
			return p.LandmarkCards[name]
		},
	}
}

func newLandmarkMaxPrereq(max int) prereq {
	return prereq{
		Desc: fmt.Sprintf("If player has the less than %d constructed landmarks (excluding City Hall). ", max+1),
		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) bool {
			if rlr.LandmarkCards["City Hall"] {
				return len(rlr.LandmarkCards)-1 < max
			}

			return len(rlr.LandmarkCards) < max
		},
	}
}

func newLandmarkMinPrereq(min int) prereq {
	return prereq{
		Desc: fmt.Sprintf("If player has the more than %d constructed landmarks (excluding City Hall). ", min+1),
		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) bool {
			if rlr.LandmarkCards["City Hall"] {
				return len(rlr.LandmarkCards)-1 > min
			}

			return len(rlr.LandmarkCards) > min
		},
	}
}

func newRollerPayout(payout int) effect {
	return newRollerPayoutWithPrereq(payout, nullPrereq)
}

func newIconCardPayout(payout int, icon string) effect {
	return effect{
		Priority: 1,

		Description: func() string {
			return fmt.Sprintf("Get %d coins from the bank for each [%s] establishment that you own on your turn only", payout, icon)
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			if p != rlr {
				return
			}

			iconCards := market.FindByIcon(icon)
			iconCardCount := 0
			for _, iconCard := range iconCards {
				iconCardCount += p.SupplyCards[iconCard.Name]
			}
			totalPayout := landmarkCardAgumentedPayout(payout, card, p) * iconCardCount * c

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &rlr.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}
}

func newCardPayout(payout int, cardName string) effect {
	return effect{
		Priority: 1,

		Description: func() string {
			return fmt.Sprintf("Get %d coins from the bank for each [%s] establishment that you own on your turn only", payout, cardName)
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			if p != rlr {
				return
			}

			cardCount := p.SupplyCards[cardName]
			totalPayout := landmarkCardAgumentedPayout(payout, card, p) * cardCount * c

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &rlr.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}
}

func counterClockwise(all []*player, rlr *player) []*player {
	var reversed []*player
	index := rlr.ID

	for i := 0; i < len(all); i++ {
		reversed = append(reversed, all[(len(all)+index-i)%len(all)])
	}

	return reversed
}
