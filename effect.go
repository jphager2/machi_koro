package main

import (
	"fmt"
)

type effect struct {
	Priority    int
	Description func() string
	Call        func(card supplyCard, rlr *player, p *player, c int)
}

func landmarkCardAgumentedPayout(payout int, card supplyCard, p *player) int {
	if p.LandmarkCards["Shopping Mall"] && (card.Icon == "Cup" || card.Icon == "Bread") {
		return payout + 1
	}
	return payout
}

func newBankPayout(payout int, onlyCurrent bool) effect {
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
		Priority: 0,

		Description: func() string {
			return fmt.Sprintf("Get %s from the bank on %s", coins, receiverName)
		},

		Call: func(card supplyCard, rlr *player, p *player, c int) {
			if onlyCurrent && p != rlr {
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
	return newBankPayout(r, false)
}

func newRollerBankPayout(r int) effect {
	return newBankPayout(r, true)
}

func newRollerPayout(payout int) effect {
	var coins string

	if payout == 1 {
		coins = "1 coin"
	} else {
		coins = fmt.Sprintf("%d coins", payout)
	}

	return effect{
		Priority: 2,

		Description: func() string {
			return fmt.Sprintf("Get %s from the player who rolled the dice", coins)
		},

		Call: func(card supplyCard, rlr *player, p *player, c int) {
			if p == rlr {
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

func newIconCardPayout(payout int, icon string) effect {
	return effect{
		Priority: 0,

		Description: func() string {
			return fmt.Sprintf("Get %d coins from the bank for each [%s] establishment that you own on your turn only", payout, icon)
		},

		Call: func(card supplyCard, rlr *player, p *player, c int) {
			if p != rlr {
				return
			}

			iconCards := supplyCards.FindByIcon(icon)
			iconCardCount := 0
			for _, iconCard := range iconCards {
				iconCardCount += p.SupplyCards[iconCard.Name]
			}
			totalPayout := payout * iconCardCount * c

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
