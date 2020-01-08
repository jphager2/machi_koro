package main

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	supplyCards supplyCardCollection

	bank = coinSet{
		OneCoins:  42,
		FiveCoins: 24,
		TenCoins:  12,
	}

	stadiumEffect = effect{
		Priority: 1,

		Description: func() string {
			return "Get 2 coins from all players on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int) {
			totalPayout := 2 * c

			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)

			for _, plr := range plrs {
				if plr == rlr {
					continue
				}

				fmt.Printf("Player %d gets %d coins from player %d [%s]\n", rlr.ID, totalPayout, plr.ID, card.Name)
				remainder := plr.Coins.TransferTo(totalPayout, &rlr.Coins)

				if remainder > 0 {
					fmt.Printf("Player %d did not have enough money. Missing: %d\n", plr.ID, remainder)
				}
			}
		},
	}

	tvStationEffect = effect{
		Priority: 1,

		Description: func() string {
			return "Take 5 coins from any one player on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int) {
			var choices []int
			var choice int
			var err error

			totalPayout := 5 * c

			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)
			fmt.Println("Pick a player to take coins from: ")

			for _, plr := range plrs {
				if plr == rlr {
					continue
				}

				choices = append(choices, plr.ID)
				fmt.Printf("Player (%d) has %d coins\n", plr.ID, plr.Coins.Total())
			}

			for {
				choice, err = scanInt(choices)

				if err != nil {
					fmt.Println(err)
					continue
				}

				break
			}

			plr := plrs[choice]

			fmt.Printf("Player %d gets %d coins from player %d [%s]\n", rlr.ID, totalPayout, plr.ID, card.Name)
			remainder := plr.Coins.TransferTo(totalPayout, &rlr.Coins)

			if remainder > 0 {
				fmt.Printf("Player %d did not have enough money. Missing: %d\n", plr.ID, remainder)
			}

			return
		},
	}

	businessCenterEffect = effect{
		Priority: 2,

		Description: func() string {
			return "Trade one non major establishment with any one player on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int) {
			fmt.Print(card.Effect.Description())
			fmt.Printf(" [%s]\n", card.Name)

			for i := 0; i < c; i++ {
				plrChoices := []int{}
				cardChoices := make(map[int][]int)
				cardChoiceNames := make(map[int][]string)

				for _, plr := range plrs {
					if plr == rlr {
						fmt.Printf("Roller has cards:\n")
					} else {
						plrChoices = append(plrChoices, plr.ID)
						fmt.Printf("Player (%d) has cards:\n", plr.ID)
					}

					j := 1
					for cardName, cardCount := range plr.SupplyCards {
						currentCard := supplyCards.FindByName(cardName)

						if currentCard.Icon == "Major" || cardCount == 0 {
							continue
						}
						cardChoices[plr.ID] = append(cardChoices[plr.ID], j)
						cardChoiceNames[plr.ID] = append(cardChoiceNames[plr.ID], cardName)
						fmt.Printf("  (%d) %s [%d]\n", j, cardName, cardCount)
						j++
					}
				}

				fmt.Println("Pick a player to trade cards with: ")
				plrID, err := scanInt(plrChoices)
				if err != nil {
					fmt.Println(err)
					continue
				}

				fmt.Println("Pick a card to take: ")
				takeCardIdx, err := scanInt(cardChoices[plrID])
				if err != nil {
					fmt.Println(err)
					continue
				}
				takeCardName := cardChoiceNames[plrID][takeCardIdx-1]

				fmt.Println("Pick a card to give: ")
				giveCardIdx, err := scanInt(cardChoices[rlr.ID])
				if err != nil {
					fmt.Println(err)
					continue
				}
				giveCardName := cardChoiceNames[rlr.ID][giveCardIdx-1]

				fmt.Printf("Player %d trades %s for %s with player %d [%s]\n", rlr.ID, giveCardName, takeCardName, plrID, card.Name)
				rlr.SupplyCards[giveCardName]--
				rlr.SupplyCards[takeCardName]++

				for _, plr := range plrs {
					if plr.ID != plrID {
						continue
					}

					p.SupplyCards[takeCardName]--
					p.SupplyCards[giveCardName]++
				}
			}
		},
	}

	landmarkCardsSorted []landmarkCard
	landmarkCards       map[string]landmarkCard

	plrs []*player
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	supplyCards = newSupplyCardCollection(
		[]*supplyCard{
			&supplyCard{
				Name:          "Wheat Field",
				Cost:          1,
				ActiveNumbers: []int{1},
				Effect:        newAllBankPayout(1),
				Icon:          "Wheat",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Ranch",
				Cost:          1,
				ActiveNumbers: []int{2},
				Effect:        newAllBankPayout(1),
				Icon:          "Cow",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Bakery",
				Cost:          1,
				ActiveNumbers: []int{2, 3},
				Effect:        newRollerBankPayout(1),
				Icon:          "Bread",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Cafe",
				Cost:          2,
				ActiveNumbers: []int{3},
				Effect:        newRollerPayout(1),
				Icon:          "Cup",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Convenience Store",
				Cost:          2,
				ActiveNumbers: []int{4},
				Effect:        newRollerBankPayout(3),
				Icon:          "Bread",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Forest",
				Cost:          3,
				ActiveNumbers: []int{5},
				Effect:        newAllBankPayout(1),
				Icon:          "Gear",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Stadium",
				Cost:          6,
				ActiveNumbers: []int{6},
				Effect:        stadiumEffect,
				Icon:          "Major",
				Supply:        4,
			},
			&supplyCard{
				Name:          "TV Station",
				Cost:          7,
				ActiveNumbers: []int{6},
				Effect:        tvStationEffect,
				Icon:          "Major",
				Supply:        4,
			},
			&supplyCard{
				Name:          "Business Center",
				Cost:          8,
				ActiveNumbers: []int{6},
				Effect:        businessCenterEffect,
				Icon:          "Major",
				Supply:        4,
			},
			&supplyCard{
				Name:          "Cheese Factory",
				Cost:          5,
				ActiveNumbers: []int{7},
				Effect:        newIconCardPayout(3, "Cow"),
				Icon:          "Factory",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Furniture Factory",
				Cost:          3,
				ActiveNumbers: []int{8},
				Effect:        newIconCardPayout(3, "Gear"),
				Icon:          "Factory",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Mine",
				Cost:          6,
				ActiveNumbers: []int{9},
				Effect:        newAllBankPayout(5),
				Icon:          "Gear",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Family Restaurant",
				Cost:          3,
				ActiveNumbers: []int{9, 10},
				Effect:        newRollerPayout(2),
				Icon:          "Cup",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Apple Orchard",
				Cost:          3,
				ActiveNumbers: []int{10},
				Effect:        newAllBankPayout(3),
				Icon:          "Wheat",
				Supply:        6,
			},
			&supplyCard{
				Name:          "Fruit and Vegetable Market",
				Cost:          2,
				ActiveNumbers: []int{11, 12},
				Effect:        newIconCardPayout(2, "Wheat"),
				Icon:          "Fruit",
				Supply:        6,
			},
		},
	)

	landmarkCardsSorted = []landmarkCard{
		landmarkCard{
			Name:        "Train Station",
			Cost:        4,
			Description: "You may roll 1 or 2 dice",
		},
		landmarkCard{
			Name:        "Shopping Mall",
			Cost:        10,
			Description: "Each of your [Cup] and [Bread] establishments earn +1 coin",
		},
		landmarkCard{
			Name:        "Amusement Park",
			Cost:        16,
			Description: "If you roll doubles take another turn after this one",
		},
		landmarkCard{
			Name:        "Radio Tower",
			Cost:        22,
			Description: "Once every turn you can choose to re-roll your dice",
		},
	}
	landmarkCards = make(map[string]landmarkCard)
	for _, card := range landmarkCardsSorted {
		landmarkCards[card.Name] = card
	}
}
