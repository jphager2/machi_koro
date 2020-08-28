package main

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	market marketplace

	bank = coinSet{
		OneCoins:  42,
		FiveCoins: 24,
		TenCoins:  12,
	}

	renovationCompanyEffect = effect{
		Priority: 1,

		Description: func() string {
			return "Choose a non-[Major] building. All buildings owned by any player of that type are closed for renovations. Get 1 coin from the bank for each building closed for renovation, on your turn only."
		},

		// TODO
		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			totalPayout := specialRoll * c

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}

	demolitionCompanyEffect = effect{
		Priority: 1,

		Description: func() string {
			return "For each Demolition Company you own, you must demolish a constructed landmark and take 8 coins from the bank"
		},

		// TODO
		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			totalPayout := specialRoll * c

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}

	tunaBoatEffect = effect{
		Priority: 1,

		Description: func() string {
			return "If you have the [Harbor] landmark. Roller rolls 2 dice and you receive that many coins from the bank on anyone's turn"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			if !p.LandmarkCards["Harbor"] {
				return
			}

			totalPayout := specialRoll * c

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}

	taxOfficeEffect = effect{
		Priority: 2,

		Description: func() string {
			return "For each players with 10 or more coins, you get half of their coins on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			for _, plr := range plrs {
				if plr == rlr || plr.Coins.Total() < 10 {
					continue
				}
				totalPayout := plr.Coins.Total() / 2

				fmt.Printf("Player %d gets %d coins from player %d [%s]\n", rlr.ID, totalPayout, plr.ID, card.Name)
				remainder := plr.Coins.TransferTo(totalPayout, &rlr.Coins)

				if remainder > 0 {
					fmt.Printf("Player %d did not have enough money. Missing: %d\n", plr.ID, remainder)
				}
			}
		},
	}

	publisherEffect = effect{
		Priority: 2,

		Description: func() string {
			return "Get 1 coin from each player for each [Cup] and [Bread] they have on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
			for _, plr := range plrs {
				if plr == rlr {
					continue
				}

				iconCards := append(market.FindByIcon("Cup"), market.FindByIcon("Bread")...)
				iconCardCount := 0
				for _, iconCard := range iconCards {
					iconCardCount += plr.SupplyCards[iconCard.Name]
				}
				totalPayout := 1 * iconCardCount * c

				fmt.Printf("Player %d gets %d coins from player %d [%s]\n", rlr.ID, totalPayout, plr.ID, card.Name)
				remainder := plr.Coins.TransferTo(totalPayout, &rlr.Coins)

				if remainder > 0 {
					fmt.Printf("Player %d did not have enough money. Missing: %d\n", plr.ID, remainder)
				}
			}
		},
	}

	stadiumEffect = effect{
		Priority: 2,

		Description: func() string {
			return "Get 2 coins from each player on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
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
		Priority: 2,

		Description: func() string {
			return "Take 5 coins from any one player on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
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

		Call: func(card supplyCard, rlr *player, p *player, c int, specialRoll int) {
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
						currentCard := market.FindByName(cardName)

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

	gameVersionsSorted []gameVersion

	plrs []*player

	cityHall = landmarkCard{
		Name:        "City Hall",
		Cost:        0,
		Description: "If you have no coins before your building phase, you may take 1 coin from the bank",
	}
	harbor = landmarkCard{
		Name:        "Harbor",
		Cost:        2,
		Description: "If you roll 10 or higher, you may add 2 to your roll",
	}
	trainStation = landmarkCard{
		Name:        "Train Station",
		Cost:        4,
		Description: "You may roll 1 or 2 dice",
	}
	shoppingMall = landmarkCard{
		Name:        "Shopping Mall",
		Cost:        10,
		Description: "Each of your [Cup] and [Bread] establishments earn +1 coin",
	}
	amusementPark = landmarkCard{
		Name:        "Amusement Park",
		Cost:        16,
		Description: "If you roll doubles take another turn after this one",
	}
	radioTower = landmarkCard{
		Name:        "Radio Tower",
		Cost:        22,
		Description: "Once every turn you can choose to re-roll your dice",
	}
	airport = landmarkCard{
		Name:        "Airport",
		Cost:        30,
		Description: "If you do not build on your turn, you may take 10 coins from the bank",
	}

	lessThanTwoLandmarksPrereq = newLandmarkMaxPrereq(2)
	moreThanTwoLandmarksPrereq = newLandmarkMinPrereq(1)

	basicSupplyCards = []*supplyCard{
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
	}

	harborSupplyCards = []*supplyCard{
		&supplyCard{
			Name:          "Pizza Joint",
			Cost:          1,
			ActiveNumbers: []int{7},
			Effect:        newRollerPayout(1),
			Icon:          "Cup",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Tax Office",
			Cost:          4,
			ActiveNumbers: []int{8, 9},
			Effect:        taxOfficeEffect,
			Icon:          "Major",
			Supply:        4,
		},
		&supplyCard{
			Name:          "Hamburger Stand",
			Cost:          1,
			ActiveNumbers: []int{8},
			Effect:        newRollerPayout(1),
			Icon:          "Cup",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Sushi Bar",
			Cost:          1,
			ActiveNumbers: []int{1},
			Effect:        newRollerPayoutWithPrereq(1, newLandmarkPrereq("Harbor", false)),
			Icon:          "Cup",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Flower Garden",
			Cost:          2,
			ActiveNumbers: []int{4},
			Effect:        newAllBankPayout(1),
			Icon:          "Wheat",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Flower Shop",
			Cost:          1,
			ActiveNumbers: []int{2},
			Effect:        newCardPayout(1, "Flower Garden"),
			Icon:          "Bread",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Food Warehouse",
			Cost:          2,
			ActiveNumbers: []int{12, 13},
			Effect:        newIconCardPayout(2, "Cup"),
			Icon:          "Factory",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Mackerel Boat",
			Cost:          2,
			ActiveNumbers: []int{8},
			Effect:        newAllBankPayout(2),
			Icon:          "Boat",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Publisher",
			Cost:          5,
			ActiveNumbers: []int{7},
			Effect:        publisherEffect,
			Icon:          "Major",
			Supply:        4,
		},
		&supplyCard{
			Name:          "Tuna Boat",
			Cost:          5,
			ActiveNumbers: []int{12, 13, 14},
			Effect:        tunaBoatEffect,
			Icon:          "Boat",
			Supply:        6,
		},
	}

	millionaireSupplyCards = []*supplyCard{
		&supplyCard{
			Name:          "General Store",
			Cost:          0,
			ActiveNumbers: []int{2},
			Effect:        newRollerBankPayoutWithPrereq(2, lessThanTwoLandmarksPrereq),
			Icon:          "Bread",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Corn Field",
			Cost:          2,
			ActiveNumbers: []int{3, 4},
			Effect:        newAllBankPayoutWithPrereq(1, lessThanTwoLandmarksPrereq),
			Icon:          "Wheat",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Demolition Company",
			Cost:          2,
			ActiveNumbers: []int{4},
			Effect:        demolitionCompanyEffect,
			Icon:          "Suitcase",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Loan Office",
			Cost:          -5,
			ActiveNumbers: []int{5, 6},
			Effect:        newBankRollerPayout(2),
			Icon:          "Suitcase",
			Supply:        6,
		},
		&supplyCard{
			Name:          "French Restaurant",
			Cost:          3,
			ActiveNumbers: []int{5},
			Effect:        newRollerPayoutWithPrereq(5, moreThanTwoLandmarksPrereq),
			Icon:          "Cup",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Vineyard",
			Cost:          3,
			ActiveNumbers: []int{7},
			Effect:        newAllBankPayout(3),
			Icon:          "Wheat",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Renovation Company",
			Cost:          4,
			ActiveNumbers: []int{8},
			Effect:        renovationCompanyEffect,
			Icon:          "Major",
			Supply:        4,
		},
	}
)

func postInit() {
	landmarkCards = make(map[string]landmarkCard)
	for _, card := range landmarkCardsSorted {
		landmarkCards[card.Name] = card
	}
}

func initBasic() {
	market = newBasicMarketplace(basicSupplyCards)
	landmarkCardsSorted = []landmarkCard{
		trainStation,
		shoppingMall,
		amusementPark,
		radioTower,
	}
	postInit()
}

func initHarbor() {
	market = newExpansionMarketplace(append(basicSupplyCards, harborSupplyCards...))
	landmarkCardsSorted = []landmarkCard{
		cityHall,
		harbor,
		trainStation,
		shoppingMall,
		amusementPark,
		radioTower,
		airport,
	}
	postInit()
}

func initMillionaire() {
	market = newExpansionMarketplace(append(basicSupplyCards, millionaireSupplyCards...))
	landmarkCardsSorted = []landmarkCard{
		cityHall,
		trainStation,
		shoppingMall,
		amusementPark,
		radioTower,
	}
	postInit()
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	gameVersionsSorted = []gameVersion{
		gameVersion{
			Name: "Basic",
			Init: initBasic,
		},
		gameVersion{
			Name: "The Harbor",
			Init: initHarbor,
		},
		gameVersion{
			Name: "Millionaire's row",
			Init: initMillionaire,
		},
	}
}
