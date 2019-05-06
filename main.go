package main

import (
	"fmt"
)

type CoinSet struct {
	OneCoins  int
	FiveCoins int
	TenCoins  int
}

func (c *CoinSet) Total() int {
	return c.OneCoins + c.FiveCoins*5 + c.TenCoins*10
}

func (c *CoinSet) Sub(amount int) (int, int, int, int) {
	var ones int
	var fives int
	var tens int

	for amount >= 10 && c.TenCoins > 0 {
		tens += 1
		c.TenCoins -= 1
		amount -= 10
	}
	for amount >= 5 && c.FiveCoins > 0 {
		fives += 1
		c.FiveCoins -= 1
		amount -= 5
	}
	for amount >= 1 && c.OneCoins > 0 {
		ones += 1
		c.OneCoins -= 1
		amount -= 1
	}

	return ones, fives, tens, amount
}

func (c *CoinSet) Add(ones int, fives int, tens int) {
	c.OneCoins += ones
	c.FiveCoins += fives
	c.TenCoins += tens
}

func (c *CoinSet) WithdrawTo(amount int, p *Player) int {
	ones, fives, tens, remainder := c.Sub(amount)
	p.Coins.Add(ones, fives, tens)

	return remainder
}

type Player struct {
	Id          int
	SupplyCards map[string]int
	Coins       CoinSet
}

type Effect struct {
	Description func() string
	Call        func(card SupplyCard, roller *Player, all []*Player)
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
				receivers = all
			}

			for _, player := range receivers {
				cardCount := player.SupplyCards[card.Name]
				totalPayout := payout * cardCount
				if cardCount == 0 {
					continue
				}
				fmt.Print(card.Effect.Description())
				fmt.Printf(" [%s]\n", card.Name)
				remainder := bank.WithdrawTo(totalPayout, player)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}
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
			for _, player := range all {
				if player == roller {
					continue
				}
				cardCount := player.SupplyCards[card.Name]
				totalPayout := payout * cardCount
				if cardCount == 0 {
					continue
				}
				fmt.Print(card.Effect.Description())
				fmt.Printf(" [%s]\n", card.Name)

				remainder := roller.Coins.WithdrawTo(totalPayout, player)

				if remainder > 0 {
					fmt.Printf("Roller did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}
}

type SupplyCardCollection struct {
	Cards []SupplyCard
}

func (s *SupplyCardCollection) FindByRole(role int) []SupplyCard {
	var found []SupplyCard

	for _, card := range s.Cards {
		for _, number := range card.ActiveNumbers {
			if number == role {
				found = append(found, card)
				break
			}
		}
	}

	return found
}

type SupplyCard struct {
	Name          string
	Cost          int
	ActiveNumbers []int
	Effect        Effect
}

var (
	bank = CoinSet{
		OneCoins:  42,
		FiveCoins: 24,
		TenCoins:  12,
	}

	supplyCards = SupplyCardCollection{
		Cards: []SupplyCard{
			SupplyCard{
				Name:          "Wheat Field",
				Cost:          1,
				ActiveNumbers: []int{1},
				Effect:        NewBankPayout(1, false),
			},
			SupplyCard{
				Name:          "Ranch",
				Cost:          1,
				ActiveNumbers: []int{2},
				Effect:        NewBankPayout(1, false),
			},
			SupplyCard{
				Name:          "Bakery",
				Cost:          1,
				ActiveNumbers: []int{2, 3},
				Effect:        NewBankPayout(1, true),
			},
			SupplyCard{
				Name:          "Cafe",
				Cost:          2,
				ActiveNumbers: []int{3},
				Effect:        NewRollerPayout(1),
			},
		},
	}

	players []*Player
)

func Roll(roller *Player, r int) {
	cards := supplyCards.FindByRole(r)

	for _, card := range cards {
		card.Effect.Call(card, roller, players)
	}
}

func main() {
	fmt.Println("machi koro!")

	playerCount := 2

	for i := 0; i < playerCount; i++ {
		player := Player{Id: i}
		players = append(players, &player)
		remainder := bank.WithdrawTo(3, &player)

		if remainder > 0 {
			fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
		}
		player.SupplyCards = make(map[string]int)
		player.SupplyCards["Wheat Field"] += 1
		player.SupplyCards["Bakery"] += 1
		player.SupplyCards["Cafe"] += 1

		fmt.Printf("Bank: %d Coins\n", bank.Total())
		fmt.Printf("%d: %d Coins\n", i, player.Coins.Total())
		fmt.Println(player.SupplyCards)
	}

	// Roll 1 three times
	for i := 0; i < 3; i++ {
		roller := players[0]
		Roll(roller, 1)
	}

	for i, player := range players {
		fmt.Printf("%d: %d Coins\n", i, player.Coins.Total())
	}

	// Roll 2 three times
	for i := 0; i < 3; i++ {
		roller := players[0]
		Roll(roller, 2)
	}

	for i, player := range players {
		fmt.Printf("%d: %d Coins\n", i, player.Coins.Total())
	}

	// Roll 3 three times
	for i := 0; i < 3; i++ {
		roller := players[0]
		Roll(roller, 3)
	}

	for i, player := range players {
		fmt.Printf("%d: %d Coins\n", i, player.Coins.Total())
	}
}
