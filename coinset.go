package main

import "fmt"

type coinSet struct {
	OneCoins  int
	FiveCoins int
	TenCoins  int
}

func (c *coinSet) Total() int {
	return c.OneCoins + c.FiveCoins*5 + c.TenCoins*10
}

// This is super complicated, since we are talking about coins and not absolute
// numbers. But it should be more accurate for the game. However, because it is
// coins, we have to have some algorithm for trading coins to the bank. The one
// provided here doesn't seem ideal, but it should work in most cases.
func (c *coinSet) Sub(amount int) (int, int, int, int) {
	var ones int
	var fives int
	var tens int

	for amount >= 10 && c.TenCoins > 0 {
		tens++
		c.TenCoins--
		amount -= 10
	}
	for amount >= 5 && c.FiveCoins > 0 {
		fives++
		c.FiveCoins--
		amount -= 5
	}
	for amount >= 1 && c.OneCoins > 0 {
		ones++
		c.OneCoins--
		amount--
	}

	return ones, fives, tens, amount
}

func (c *coinSet) Add(ones int, fives int, tens int) {
	c.OneCoins += ones
	c.FiveCoins += fives
	c.TenCoins += tens

	_ = c.TradeWithBank()
}

func (c *coinSet) TradeWithBank() bool {
	ok := true
	var wantFives int
	var wantOnes int
	var giveFives int
	var giveOnes int

	wantFives = 2 - c.FiveCoins
	wantOnes = 5 - c.OneCoins

	if wantOnes > 0 && wantFives > 0 {
		wantFives++
	}

	if wantFives > 0 && c.TenCoins > 0 {
		if bank.FiveCoins < wantFives {
			ok = false
		} else {
			bank.FiveCoins -= 2
			c.FiveCoins += 2
			c.TenCoins--
			bank.TenCoins++
		}
	}

	if wantOnes > 0 && c.FiveCoins > 0 {
		if bank.OneCoins < wantOnes {
			ok = false
		} else {
			bank.OneCoins -= 5
			c.OneCoins += 5
			c.FiveCoins--
			bank.FiveCoins++
		}
	}

	giveOnes = c.OneCoins/5 - 1
	if giveOnes > 0 {
		if bank.FiveCoins < giveOnes {
			ok = false
		} else {
			bank.FiveCoins -= giveOnes
			c.FiveCoins += giveOnes
			c.OneCoins -= 5 * giveOnes
			bank.OneCoins += 5 * giveOnes
		}
	}

	giveFives = c.FiveCoins/2 - 1
	if giveFives > 0 {
		if bank.TenCoins < giveFives {
			ok = false
		} else {
			bank.TenCoins -= giveFives
			c.TenCoins += giveFives
			c.OneCoins -= 2 * giveFives
			bank.OneCoins += 2 * giveFives
		}
	}

	return ok
}

func (c *coinSet) TransferTo(amount int, receiver *coinSet) int {
	ones, fives, tens, remainder := c.Sub(amount)

	receiver.Add(ones, fives, tens)

	// If there is a remainder less than the total of the player's coins, that
	// means that they have a coin that is larger, and it needs to be broken down
	// by the bank. Also coin sets should trade up and down with the bank so that
	// if any denomination is greater than the next largest denomination, the bank
	// will trade up, and if it is less than or equal to it, it should trade down.
	// For example if use has (ones, fives, tens : 5 5 0) he should trade 2 fives
	// to the bank for 1 ten. If a player has (0, 1, 1) He should trade 1 five for
	// 5 ones, and 1 ten for 2 fives.
	if remainder > 0 && remainder <= c.Total() {
		ok := c.TradeWithBank()
		if !ok {
			fmt.Println("Failed to trade coins with bank.")
			return remainder
		}
		return c.TransferTo(remainder, receiver)
	}

	return remainder
}
