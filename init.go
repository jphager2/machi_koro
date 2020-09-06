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

	atLeastThreeLandmarks = newLandmarkMinPrereq(2)
	membersOnlyClubEffect = effect{
		Priority: 1,

		Description: func() string {
			return atLeastThreeLandmarks.Desc + "If the player who rolled the dice has 3 or more constructed landmarks, get all of their coins."
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p == rlr {
				return
			}
			if !atLeastThreeLandmarks.Call(card, rlr, p, c, pc, specialRoll) {
				return
			}

			totalPayout := rlr.Coins.Total()

			fmt.Printf("Player %d gets %d coins from the player %d [%s].\n", p.ID, totalPayout, rlr.ID, card.Name)
			remainder := bank.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Roller did not have enough money. Missing: %d\n", remainder)
			}
		},
	}

	parkEffect = effect{
		Priority: 2,

		Description: func() string {
			return "Redistribute all players' coins evenly among all players (if there is an uneven amount of coins, take coins from the bank to make up the difference), on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

			var hoard coinSet

			for _, plr := range plrs {
				fmt.Printf("Player %d puts %d coins into the redistribution.\n", p.ID, plr.Coins.Total())
				plr.Coins.TransferTo(plr.Coins.Total(), &hoard)
			}

			totalPayout := hoard.Total() / len(plrs)
			missing := hoard.Total() % len(plrs)
			remainder := bank.TransferTo(missing, &hoard)
			if remainder > 0 {
				fmt.Printf("Bank did not have enough money to fill up the redistribution. Missing: %d\n", remainder)
			}
			// If there is missing coins, total payout is 1 short
			if missing > 0 {
				totalPayout++
			}

			for _, plr := range counterClockwise(plrs, rlr) {
				fmt.Printf("Player %d gets %d coins from the redistribution [%s].\n", p.ID, totalPayout, card.Name)
				remainder = hoard.TransferTo(totalPayout, &plr.Coins)

				if remainder > 0 {
					fmt.Printf("Redistribution did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}

	sodaBottlingPlantEffect = effect{
		Priority: 1,

		Description: func() string {
			return "Get 1 coin from the bank for every [Cup] owned by all players, on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

			iconCardCount := 0
			iconCards := market.FindByIcon("Cup")

			for _, plr := range plrs {
				for _, iconCard := range iconCards {
					iconCardCount += plr.SupplyCards[iconCard.Name].Total
				}
			}

			totalPayout := 1 * iconCardCount

			fmt.Printf("Player %d gets %d coins from the bank [%s].\n", p.ID, totalPayout, card.Name)
			remainder := bank.TransferTo(totalPayout, &p.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		},
	}

	// NOTE: While money is put on each card, the money is actually pooled, since Major cards cannot be closed for renovation or traded.
	techStartupEffect = effect{
		Priority: 2,

		Description: func() string {
			return "At the end of your turn you can put 1 coin on this card. If this card is activated, you get that many coins from each player, on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

			totalPayout := p.Investment.Total()

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

	wineryPayout = newCardPayout(6, "Vineyard")
	wineryEffect = effect{
		Priority: 1,

		Description: func() string {
			return "Get 6 coins for each vineyard you have, then close this building for renovation, on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

			wineryPayout.Call(card, rlr, p, c, pc, specialRoll)

			fmt.Printf("%d of Player %d's %s cards are closed for renovation.\n", c, p.ID, card.Name)

			pc.Renovation = pc.Total
		},
	}

	movingCompanyEffect = effect{
		Priority: 1,

		Description: func() string {
			return "You must give a non-[Major] building you own to another player. When you do, get 4 coins from the bank, on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

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
					for cardName, playerCard := range plr.SupplyCards {
						currentCard := market.FindByName(cardName)

						if currentCard.Icon == "Major" || playerCard.Total == 0 {
							continue
						}
						cardChoices[plr.ID] = append(cardChoices[plr.ID], j)
						cardChoiceNames[plr.ID] = append(cardChoiceNames[plr.ID], cardName)
						fmt.Printf("  (%d) %s [%d]\n", j, cardName, playerCard.Total)
						j++
					}
				}

				var plrID int
				var giveCardIdx int
				var err error

				for {
					fmt.Println("Pick a player to trade cards with: ")
					plrID, err = scanInt(plrChoices)
					if err != nil {
						fmt.Println(err)
						continue
					}
					break
				}

				for {
					fmt.Println("Pick a card to give: ")
					giveCardIdx, err = scanInt(cardChoices[rlr.ID])
					if err != nil {
						fmt.Println(err)
						continue
					}
					break
				}
				giveCardName := cardChoiceNames[rlr.ID][giveCardIdx-1]

				fmt.Printf("Player %d gives %s to player %d [%s]\n", rlr.ID, giveCardName, plrID, card.Name)
				rlr.SupplyCards[giveCardName].Total--

				for _, plr := range plrs {
					if plr.ID != plrID {
						continue
					}

					p.SupplyCards[giveCardName].Total++
				}

				fmt.Printf("Player %d gets 4 coins from the bank [%s]\n", rlr.ID, card.Name)
				remainder := bank.TransferTo(4, &rlr.Coins)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}

	renovationCompanyEffect = effect{
		Priority: 2,

		Description: func() string {
			return "Choose a non-[Major] building. All buildings owned by any player of that type are closed for renovations. Get 1 coin from the bank for each building closed for renovation, on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

			for i := 0; i < c; i++ {
				plrChoices := []int{}
				cardChoicesCount := make(map[string]int)

				for _, plr := range plrs {
					if plr == rlr {
						fmt.Printf("Roller has cards:\n")
					} else {
						plrChoices = append(plrChoices, plr.ID)
						fmt.Printf("Player (%d) has cards:\n", plr.ID)
					}

					j := 1
					for cardName, playerCard := range plr.SupplyCards {
						currentCard := market.FindByName(cardName)

						if currentCard.Icon == "Major" || playerCard.Total == 0 {
							continue
						}
						cardChoicesCount[cardName]++
						j++
					}
				}

				var cardIdx int
				var cardChoices []int
				var cardChoiceNames []string
				var err error
				var i int

				for cardName, count := range cardChoicesCount {
					cardChoices = append(cardChoices, i)
					cardChoiceNames = append(cardChoiceNames, cardName)
					fmt.Printf("  (%d) %s [%d]\n", i, cardName, count)
					i++
				}
				fmt.Println("Pick a card to close for renovation: ")
				cardIdx, err = scanInt(cardChoices)
				if err != nil {
					fmt.Println(err)
					continue
				}
				cardName := cardChoiceNames[cardIdx-1]

				fmt.Printf("Player %d closes %s for renovations [%s]\n", rlr.ID, cardName, card.Name)

				for _, plr := range plrs {
					plr.SupplyCards[cardName].Renovation = plr.SupplyCards[cardName].Total
				}

				totalPayment := cardChoicesCount[cardName]

				fmt.Printf("Player %d gets %d coins from the bank [%s]\n", rlr.ID, totalPayment, card.Name)
				remainder := bank.TransferTo(totalPayment, &rlr.Coins)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}

	atLeastOneLandmarkPrereq = newLandmarkMinPrereq(0)
	demolitionCompanyEffect  = effect{
		Priority: 1,

		Description: func() string {
			return "For each Demolition Company you own, you must demolish a constructed landmark and take 8 coins from the bank, on your turn only"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}
			for i := 0; i < c; i++ {
				if !atLeastOneLandmarkPrereq.Call(card, rlr, p, c, pc, specialRoll) {
					return
				}

				j := 0
				choices := []int{}
				choiceNames := []string{}
				fmt.Printf("Player %d Landmarks: \n", rlr.ID)
				for _, landmark := range landmarkCardsSorted {
					if !rlr.LandmarkCards[landmark.Name] || landmark.Name == "City Hall" {
						continue
					}

					j++
					choices = append(choices, j)
					choiceNames = append(choiceNames, landmark.Name)
					fmt.Printf("  (%d) %s [%d coins]: %s\n", j, landmark.Name, landmark.Cost, landmark.Description)
				}

				var landmarkIdx int
				var err error

				for {
					fmt.Print("Which landmark do you want to demolish? ")
					landmarkIdx, err = scanInt(choices)
					if err != nil {
						fmt.Println("No landmark selected.")
						continue
					}
					break
				}
				landmarkName := choiceNames[landmarkIdx-1]
				rlr.LandmarkCards[landmarkName] = false

				fmt.Printf("Player %d gets 8 coins from the bank [%s].\n", rlr.ID, card.Name)
				remainder := bank.TransferTo(8, &rlr.Coins)

				if remainder > 0 {
					fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
				}
			}
		},
	}

	tunaBoatEffect = effect{
		Priority: 1,

		Description: func() string {
			return "If you have the [Harbor] landmark. Roller rolls 2 dice and you receive that many coins from the bank on anyone's turn"
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
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

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

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

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

			for _, plr := range plrs {
				if plr == rlr {
					continue
				}

				iconCards := append(market.FindByIcon("Cup"), market.FindByIcon("Bread")...)
				iconCardCount := 0
				for _, iconCard := range iconCards {
					iconCardCount += plr.SupplyCards[iconCard.Name].Total
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

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

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

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

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
			return "Trade one non major establishment with any one player on your turn only."
		},

		Call: func(card supplyCard, rlr *player, p *player, c int, pc *playerCard, specialRoll int) {
			if p != rlr {
				return
			}

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
					for cardName, playerCard := range plr.SupplyCards {
						currentCard := market.FindByName(cardName)

						if currentCard.Icon == "Major" || playerCard.Total == 0 {
							continue
						}
						cardChoices[plr.ID] = append(cardChoices[plr.ID], j)
						cardChoiceNames[plr.ID] = append(cardChoiceNames[plr.ID], cardName)
						fmt.Printf("  (%d) %s [%d]\n", j, cardName, playerCard.Total)
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
				rlr.SupplyCards[giveCardName].Total--
				rlr.SupplyCards[takeCardName].Total++

				for _, plr := range plrs {
					if plr.ID != plrID {
						continue
					}

					p.SupplyCards[takeCardName].Total--
					p.SupplyCards[giveCardName].Total++
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
		&supplyCard{
			Name:          "Moving Company",
			Cost:          2,
			ActiveNumbers: []int{9, 10},
			Effect:        movingCompanyEffect,
			Icon:          "Suitcase",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Winery",
			Cost:          3,
			ActiveNumbers: []int{9},
			Effect:        wineryEffect,
			Icon:          "Factory",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Tech Startup",
			Cost:          1,
			ActiveNumbers: []int{10},
			Effect:        techStartupEffect,
			Icon:          "Major",
			Supply:        4,
		},
		&supplyCard{
			Name:          "Soda Bottling Plant",
			Cost:          5,
			ActiveNumbers: []int{11},
			Effect:        sodaBottlingPlantEffect,
			Icon:          "Factory",
			Supply:        6,
		},
		&supplyCard{
			Name:          "Park",
			Cost:          3,
			ActiveNumbers: []int{11, 12, 13},
			Effect:        parkEffect,
			Icon:          "Major",
			Supply:        4,
		},
		&supplyCard{
			Name:          "Member's Only Club",
			Cost:          4,
			ActiveNumbers: []int{12, 13, 14},
			Effect:        membersOnlyClubEffect,
			Icon:          "Cup",
			Supply:        6,
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
