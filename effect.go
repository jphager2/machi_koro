package main

import (
	"fmt"
)

type Effect struct {
	Description func() string
	Call        func(card SupplyCard, roller *Player, all []*Player)
}

func landmarkCardAgumentedPayout(payout int, card SupplyCard, p *Player) int {
	if p.LandmarkCards["Shopping Mall"] && (card.Icon == "Cup" || card.Icon == "Bread") {
		return payout + 1
	} else {
		return payout
	}
}

func NewBankPayout(payout int, onlyCurrent bool) Effect {
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

	return Effect{
		Description: func() string {
			return fmt.Sprintf("Get %s from the bank on %s", coins, receiverName)
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
			var receivers []*Player

			if onlyCurrent {
				receivers = []*Player{roller}
			} else {
				receivers = counterClockwise(all, roller)
			}

			// fmt.Print(card.Effect.Description())
			// fmt.Printf(" [%s]\n", card.Name)

			for _, player := range receivers {
				cardCount := player.SupplyCards[card.Name]
				totalPayout := landmarkCardAgumentedPayout(payout, card, player) * cardCount
				if cardCount == 0 {
					continue
				}
				fmt.Printf("Player %d gets %d coins from the bank [%s].\n", player.ID, totalPayout, card.Name)
				remainder := bank.WithdrawTo(totalPayout, &player.Coins)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}
}

func NewAllBankPayout(r int) Effect {
	return NewBankPayout(r, false)
}

func NewRollerBankPayout(r int) Effect {
	return NewBankPayout(r, true)
}

func NewRollerPayout(payout int) Effect {
	var coins string

	if payout == 1 {
		coins = "1 coin"
	} else {
		coins = fmt.Sprintf("%d coins", payout)
	}

	return Effect{
		Description: func() string {
			return fmt.Sprintf("Get %s from the player who rolled the dice", coins)
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
			// fmt.Print(card.Effect.Description())
			// fmt.Printf(" [%s]\n", card.Name)

			receivers := counterClockwise(all, roller)
			for _, player := range receivers {
				if player == roller {
					continue
				}
				cardCount := player.SupplyCards[card.Name]
				totalPayout := landmarkCardAgumentedPayout(payout, card, player) * cardCount
				if cardCount == 0 {
					continue
				}
				fmt.Printf("Player %d gets %d coins from the player %d [%s].\n", player.ID, totalPayout, roller.ID, card.Name)
				remainder := roller.Coins.WithdrawTo(totalPayout, &player.Coins)

				if remainder > 0 {
					fmt.Printf("Roller did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}
}

func NewIconCardPayout(payout int, icon string) Effect {
	return Effect{
		Description: func() string {
			return fmt.Sprintf("Get %d coins from the bank for each [%s] establishment that you own on your turn only", payout, icon)
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
			// fmt.Print(card.Effect.Description())
			// fmt.Printf(" [%s]\n", card.Name)

			cardCount := roller.SupplyCards[card.Name]
			iconCards := supplyCards.FindByIcon(icon)
			iconCardCount := 0
			for _, iconCard := range iconCards {
				iconCardCount += roller.SupplyCards[iconCard.Name]
			}
			totalPayout := payout * iconCardCount * cardCount

			if cardCount == 0 {
				return
			}

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", roller.ID, totalPayout, card.Name)
			remainder := bank.WithdrawTo(totalPayout, &roller.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}
}

func counterClockwise(all []*Player, roller *Player) []*Player {
	var reversed []*Player
	index := roller.ID

	for i := 0; i < len(all); i++ {
		reversed = append(reversed, all[(len(all)+index-i)%len(all)])
	}

	return reversed
}
