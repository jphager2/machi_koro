package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func NewIconCardPayout(payout int, icon string) Effect {
	return Effect{
		Description: func() string {
			return fmt.Sprintf("Get %d coins from the bank for each [%s] establishment that you own on your turn only", payout, icon)
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
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

			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)

			remainder := bank.WithdrawTo(totalPayout, roller)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}
}

type SupplyCardCollection struct {
	Cards []SupplyCard
}

func (s *SupplyCardCollection) FindByIcon(icon string) []SupplyCard {
	var found []SupplyCard

	for _, card := range s.Cards {
		if card.Icon == icon {
			found = append(found, card)
		}
	}

	return found
}

func (s *SupplyCardCollection) FindByName(name string) SupplyCard {
	for _, card := range s.Cards {
		if name == card.Name {
			return card
		}
	}

	return SupplyCard{}
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
	Icon          string
}

func Roll(roller *Player, r int) {
	cards := supplyCards.FindByRole(r)

	for _, card := range cards {
		card.Effect.Call(card, roller, players)
	}
}

func JoinInts(ints []int, delimiter string) string {
	str := make([]string, len(ints))
	for i, v := range ints {
		str[i] = strconv.Itoa(v)
	}

	return strings.Join(str, delimiter)
}

func GetInt(validValues []int) (int, error) {
	var val int
	fmt.Scan(&val)

	for _, v := range validValues {
		if v == val {
			return val, nil
		}
	}

	return 0, errors.New(fmt.Sprintf("Invalid input '%d' for values (%s)", val, JoinInts(validValues, ",")))
}

var (
	supplyCards SupplyCardCollection

	bank = CoinSet{
		OneCoins:  42,
		FiveCoins: 24,
		TenCoins:  12,
	}

	stadiumEffect = Effect{
		Description: func() string {
			return "Get 2 coins from all players on your turn only"
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
			cardCount := roller.SupplyCards[card.Name]
			totalPayout := 2 * cardCount
			if cardCount == 0 {
				return
			}
			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)

			for _, player := range all {
				if player == roller {
					continue
				}
				remainder := player.Coins.WithdrawTo(totalPayout, roller)

				if remainder > 0 {
					fmt.Printf("Player %d did not have enough money. Missing: %d\n", player.Id, remainder)
				}
			}
		},
	}

	tvStationEffect = Effect{
		Description: func() string {
			return "Take 5 coins from any one player on your turn only"
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
			var choices []int

			cardCount := roller.SupplyCards[card.Name]
			totalPayout := 5 * cardCount
			if cardCount == 0 {
				return
			}
			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)
			fmt.Println("Pick a player to take coins from: ")

			for _, player := range all {
				if player == roller {
					continue
				}

				choices = append(choices, player.Id)
				fmt.Printf("Player %d: %d coins\n", player.Id, player.Coins.Total())
			}

			choice, err := GetInt(choices)

			if err != nil {
				fmt.Println(err)
				return
			}

			for _, player := range players {
				if player.Id != choice {
					continue
				}

				remainder := player.Coins.WithdrawTo(totalPayout, roller)

				if remainder > 0 {
					fmt.Printf("Player %d did not have enough money. Missing: %d\n", player.Id, remainder)
				}

				return
			}
		},
	}

	businessCenterEffect = Effect{
		Description: func() string {
			return "Trade one non major establishment with any one player on your turn only"
		},

		Call: func(card SupplyCard, roller *Player, all []*Player) {
			cardCount := roller.SupplyCards[card.Name]
			if cardCount == 0 {
				return
			}
			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)

			for i := 0; i < cardCount; i++ {
				playerChoices := []int{}
				cardChoices := make(map[int][]int)
				cardChoiceNames := make(map[int][]string)

				for _, player := range all {
					if player == roller {
						fmt.Printf("You have cards:\n")
					} else {
						playerChoices = append(playerChoices, player.Id)
						fmt.Printf("Player %d has cards:\n", player.Id)
					}

					j := 1
					for cardName, cardCount := range player.SupplyCards {
						currentCard := supplyCards.FindByName(cardName)

						if currentCard.Icon == "Major" || cardCount == 0 {
							continue
						}
						cardChoices[player.Id] = append(cardChoices[player.Id], j)
						cardChoiceNames[player.Id] = append(cardChoiceNames[player.Id], cardName)
						fmt.Printf("  %d) %s [%d]\n", j, cardName, cardCount)
						j++
					}
				}

				fmt.Println("Pick a player to trade cards with: ")
				playerId, err := GetInt(playerChoices)
				if err != nil {
					fmt.Println(err)
					continue
				}

				fmt.Println("Pick a card to take: ")
				takeCardIdx, err := GetInt(cardChoices[playerId])
				if err != nil {
					fmt.Println(err)
					continue
				}
				takeCardName := cardChoiceNames[playerId][takeCardIdx-1]

				fmt.Println("Pick a card to give: ")
				giveCardIdx, err := GetInt(cardChoices[roller.Id])
				if err != nil {
					fmt.Println(err)
					continue
				}
				giveCardName := cardChoiceNames[roller.Id][giveCardIdx-1]

				roller.SupplyCards[giveCardName]--
				roller.SupplyCards[takeCardName]++

				for _, player := range players {
					if player.Id != playerId {
						continue
					}

					player.SupplyCards[takeCardName]--
					player.SupplyCards[giveCardName]++
				}
			}
		},
	}

	players []*Player
)

func init() {
	supplyCards = SupplyCardCollection{
		Cards: []SupplyCard{
			SupplyCard{
				Name:          "Wheat Field",
				Cost:          1,
				ActiveNumbers: []int{1},
				Effect:        NewAllBankPayout(1),
				Icon:          "Wheat",
			},
			SupplyCard{
				Name:          "Ranch",
				Cost:          1,
				ActiveNumbers: []int{2},
				Effect:        NewAllBankPayout(1),
				Icon:          "Cow",
			},
			SupplyCard{
				Name:          "Bakery",
				Cost:          1,
				ActiveNumbers: []int{2, 3},
				Effect:        NewRollerBankPayout(1),
				Icon:          "Bread",
			},
			SupplyCard{
				Name:          "Cafe",
				Cost:          2,
				ActiveNumbers: []int{3},
				Effect:        NewRollerPayout(1),
				Icon:          "Cup",
			},
			SupplyCard{
				Name:          "Convenience Store",
				Cost:          2,
				ActiveNumbers: []int{4},
				Effect:        NewRollerBankPayout(3),
				Icon:          "Bread",
			},
			SupplyCard{
				Name:          "Forest",
				Cost:          3,
				ActiveNumbers: []int{5},
				Effect:        NewAllBankPayout(1),
				Icon:          "Gear",
			},
			SupplyCard{
				Name:          "Stadium",
				Cost:          6,
				ActiveNumbers: []int{6},
				Effect:        stadiumEffect,
				Icon:          "Major",
			},
			SupplyCard{
				Name:          "TV Station",
				Cost:          7,
				ActiveNumbers: []int{6},
				Effect:        tvStationEffect,
				Icon:          "Major",
			},
			SupplyCard{
				Name:          "Business Center",
				Cost:          8,
				ActiveNumbers: []int{6},
				Effect:        businessCenterEffect,
				Icon:          "Major",
			},
			SupplyCard{
				Name:          "Cheese Factory",
				Cost:          5,
				ActiveNumbers: []int{7},
				Effect:        NewIconCardPayout(3, "Cow"),
				Icon:          "Factory",
			},
			SupplyCard{
				Name:          "Furniture Factory",
				Cost:          3,
				ActiveNumbers: []int{8},
				Effect:        NewIconCardPayout(3, "Gear"),
				Icon:          "Factory",
			},
			SupplyCard{
				Name:          "Mine",
				Cost:          6,
				ActiveNumbers: []int{9},
				Effect:        NewAllBankPayout(5),
				Icon:          "Gear",
			},
			SupplyCard{
				Name:          "Family Restaurant",
				Cost:          3,
				ActiveNumbers: []int{9, 10},
				Effect:        NewRollerPayout(2),
				Icon:          "Cup",
			},
			SupplyCard{
				Name:          "Apple Orchard",
				Cost:          3,
				ActiveNumbers: []int{10},
				Effect:        NewAllBankPayout(3),
				Icon:          "Wheat",
			},
			SupplyCard{
				Name:          "Fruit and Vegetable Market",
				Cost:          2,
				ActiveNumbers: []int{11, 12},
				Effect:        NewIconCardPayout(2, "Wheat"),
				Icon:          "Fruit",
			},
		},
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
		player.SupplyCards["Fruit and Vegetable Market"] += 1

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

	// Roll 11 three times
	for i := 0; i < 3; i++ {
		roller := players[0]
		Roll(roller, 11)
	}

	for i, player := range players {
		fmt.Printf("%d: %d Coins\n", i, player.Coins.Total())
	}
}
