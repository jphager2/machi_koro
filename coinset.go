package main

type coinSet struct {
	OneCoins  int
	FiveCoins int
	TenCoins  int
}

func (c *coinSet) Total() int {
	return c.OneCoins + c.FiveCoins*5 + c.TenCoins*10
}

// TODO this doesn't work, right? This is too complicated anyway... just use
// absolute numbers ... or do it right, which means you should try to trade
// coins with the player or the bank...
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
}

func (c *coinSet) TransferTo(amount int, receiver *coinSet) int {
	ones, fives, tens, remainder := c.Sub(amount)
	receiver.Add(ones, fives, tens)

	return remainder
}
