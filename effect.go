package main

import (
	"fmt"
)

type effect struct {
	Description func() string
	Call        func(card supplyCard, rlr *player, all []*player)
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
		Description: func() string {
			return fmt.Sprintf("Get %s from the bank on %s", coins, receiverName)
		},

		Call: func(card supplyCard, rlr *player, all []*player) {
			var receivers []*player

			if onlyCurrent {
				receivers = []*player{rlr}
			} else {
				receivers = counterClockwise(all, rlr)
			}

			for _, p := range receivers {
				cardCount := p.SupplyCards[card.Name]
				totalPayout := landmarkCardAgumentedPayout(payout, card, p) * cardCount
				if cardCount == 0 {
					continue
				}
				fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
				remainder := bank.TransferTo(totalPayout, &p.Coins)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
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
		Description: func() string {
			return fmt.Sprintf("Get %s from the player who rolled the dice", coins)
		},

		Call: func(card supplyCard, rlr *player, all []*player) {
			receivers := counterClockwise(all, rlr)
			for _, p := range receivers {
				if p == rlr {
					continue
				}
				cardCount := p.SupplyCards[card.Name]
				totalPayout := landmarkCardAgumentedPayout(payout, card, p) * cardCount
				if cardCount == 0 {
					continue
				}
				fmt.Printf("Player %d gets %d coins from the player %d [%s].\n", p.ID, totalPayout, rlr.ID, card.Name)
				remainder := rlr.Coins.TransferTo(totalPayout, &p.Coins)

				if remainder > 0 {
					fmt.Printf("Roller did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}
}

func newIconCardPayout(payout int, icon string) effect {
	return effect{
		Description: func() string {
			return fmt.Sprintf("Get %d coins from the bank for each [%s] establishment that you own on your turn only", payout, icon)
		},

		Call: func(card supplyCard, rlr *player, all []*player) {
			cardCount := rlr.SupplyCards[card.Name]
			iconCards := supplyCards.FindByIcon(icon)
			iconCardCount := 0
			for _, iconCard := range iconCards {
				iconCardCount += rlr.SupplyCards[iconCard.Name]
			}
			totalPayout := payout * iconCardCount * cardCount

			if cardCount == 0 {
				return
			}

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", rlr.ID, totalPayout, card.Name)
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
