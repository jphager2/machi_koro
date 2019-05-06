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

type Bank struct {
	Coins CoinSet
}

func (b *Bank) WithdrawTo(amount int, p *Player) int {
	ones, fives, tens, remainder := b.Coins.Sub(amount)
	p.Coins.Add(ones, fives, tens)

	return remainder
}

type Player struct {
	SupplyCards map[string]int
	Coins       CoinSet
}

type Effect struct {
	Description func(card *SupplyCard) string
	Call        func(currentPlayer bool, roller *Player, other []*Player, b *Bank)
}

func NewBankPayout(payout int, onlyCurrent bool) Effect {
	var coins string

	if payout == 1 {
		coins = "1 coin"
	} else {
		coins = fmt.Sprintf("%d coins", payout)
	}

	return Effect{
		Description: func(c *SupplyCard) string {
			return fmt.Sprintf("Get %s from the bank on anyone's turn", coins)
		},
		Call: func(currentPlayer bool, roller *Player, other []*Player, b *Bank) {
			if !onlyCurrent || currentPlayer {
				remainder := b.WithdrawTo(payout, roller)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}
}

type SupplyCard struct {
	Name          string
	Cost          int
	ActiveNumbers []int
	Effect        Effect
}

var (
	bank = Bank{
		Coins: CoinSet{
			OneCoins:  42,
			FiveCoins: 24,
			TenCoins:  12,
		},
	}

	supplyCards = []SupplyCard{
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
	}
)

func main() {
	fmt.Println("machi koro!")

	playerCount := 2

	players := make([]Player, playerCount)

	for i, player := range players {
		remainder := bank.WithdrawTo(3, &player)

		if remainder > 0 {
			fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
		}
		player.SupplyCards = make(map[string]int)
		player.SupplyCards["Wheat Field"] += 1
		player.SupplyCards["Bakery"] += 1

		fmt.Printf("Bank: %d Coins\n", bank.Coins.Total())
		fmt.Printf("%d: %d Coins\n", i, player.Coins.Total())
		fmt.Println(player.SupplyCards)
	}

	for i := 0; i < 3; i += 1 {
		// Roll 1
		for j, player := range players {
			currentPlayer := j == 0

			var cards []SupplyCard
		}
	}
}