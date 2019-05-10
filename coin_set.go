package main

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

func (c *CoinSet) WithdrawTo(amount int, receiver *CoinSet) int {
	ones, fives, tens, remainder := c.Sub(amount)
	receiver.Add(ones, fives, tens)

	return remainder
}
