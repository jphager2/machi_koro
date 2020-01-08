package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type player struct {
	ID            int
	SupplyCards   map[string]int
	LandmarkCards map[string]bool
	Coins         coinSet
}

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

func main() {
	fmt.Println("machi koro!")

	fmt.Print("How many players (2 - 4): ")
	plrCount, err := scanInt([]int{2, 3, 4})

	if err != nil {
		fmt.Println("This game is for 2-4 players")
		os.Exit(1)
	}

	for i := 0; i < plrCount; i++ {
		p := player{ID: i}
		plrs = append(plrs, &p)
		remainder := bank.TransferTo(3, &p.Coins)

		if remainder > 0 {
			fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
		}
		p.SupplyCards = make(map[string]int)
		p.SupplyCards["Wheat Field"]++
		p.SupplyCards["Bakery"]++

		p.LandmarkCards = make(map[string]bool)
		for landmarkName := range landmarkCards {
			p.LandmarkCards[landmarkName] = false
		}
	}

	turn := 0
	reroll := false

	// Game Loop
	for {
		didAction := false
		rlr := plrs[turn]

		fmt.Printf("It's player %d's turn\n", rlr.ID)

		r, doubles, err := promptRoll(rlr.LandmarkCards["Train Station"] /* choice */)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Player %d rolls %d\n", rlr.ID, r)

		if reroll {
			reroll = false
		} else if rlr.LandmarkCards["Radio Tower"] {
			fmt.Print("Do you want to re-roll? ")

			if res := promptBool(); res {
				reroll = true
				continue
			}
		}

		// For each active card apply effects in reverse order of players.
		// Roller should get money from bank, then money from players before
		// other players receive payouts.
		cards := supplyCards.FindByRoll(r)

		for i := 0; i < len(plrs); i++ {
			p := plrs[(len(plrs)+rlr.ID-i)%len(plrs)]

			for _, card := range cards {
				c := p.SupplyCards[card.Name]
				if c > 0 {
					card.Effect.Call(*card, rlr, p, c)
				}
			}
		}

		didAction = promptSupplyCardPurchase(rlr)
		if didAction {
			didAction = false
		} else {
			didAction = promptLandmarkCardPurchase(rlr)
		}

		winner := true
		for _, hasLandmark := range rlr.LandmarkCards {
			if !hasLandmark {
				winner = false
				break
			}
		}
		if winner {
			fmt.Printf("Player %d has won the game!\n", rlr.ID)
			os.Exit(0)
		}

		if doubles && rlr.LandmarkCards["Amusement Park"] {
			continue
		}

		turn = (turn + 1) % len(plrs)
	}
}

func promptBool() bool {
	fmt.Print("(y/n) ")

	var val string
	fmt.Scan(&val)

	switch val {
	case "y", "yes":
		return true
	}
	return false
}

func scanInt(oneOf []int) (int, error) {
	var val int
	fmt.Scan(&val)

	for _, v := range oneOf {
		if v == val {
			return val, nil
		}
	}

	strInts := make([]string, len(oneOf))
	for i, v := range oneOf {
		strInts[i] = strconv.Itoa(v)
	}
	values := strings.Join(strInts, ", ")

	return 0, fmt.Errorf("Invalid input '%d' for values (%s)", val, values)
}

func promptSupplyCardPurchase(rlr *player) bool {
	fmt.Printf("Do you want to buy an establishment? (%d coins) ", rlr.Coins.Total())

	if res := promptBool(); !res {
		return false
	}

	i := 0
	choices := []int{}
	choiceNames := []string{}
	fmt.Println("Establishments: ")
	for _, card := range supplyCards.Cards {
		if card.Supply == 0 {
			continue
		}

		i++
		choices = append(choices, i)
		choiceNames = append(choiceNames, card.Name)
		fmt.Printf("  (%d) %s [%d coins] (%d left): %s\n", i, card.Name, card.Cost, card.Supply, card.Effect.Description())
	}

	fmt.Print("Which establishment do you want to buy? ")
	supplyCardIdx, err := scanInt(choices)
	if err != nil {
		fmt.Println("No establishment selected.")
		return false
	}
	supplyCardName := choiceNames[supplyCardIdx-1]
	card := supplyCards.FindByName(supplyCardName)

	if rlr.Coins.Total() < card.Cost {
		fmt.Printf("Player %d does not have enough coins to buy %s\n", rlr.ID, supplyCardName)
	} else {
		fmt.Printf("Player %d buys %s\n", rlr.ID, supplyCardName)
		rlr.Coins.TransferTo(card.Cost, &bank)
		rlr.SupplyCards[supplyCardName]++
		card.Supply--
	}
	return true
}

func promptLandmarkCardPurchase(rlr *player) bool {
	fmt.Printf("Do you want to buy a landmark? (%d coins) ", rlr.Coins.Total())

	if res := promptBool(); !res {
		return false
	}

	i := 0
	choices := []int{}
	choiceNames := []string{}
	fmt.Println("Landmarks: ")
	for _, landmark := range landmarkCardsSorted {
		if rlr.LandmarkCards[landmark.Name] {
			continue
		}

		i++
		choices = append(choices, i)
		choiceNames = append(choiceNames, landmark.Name)
		fmt.Printf("  (%d) %s [%d coins]: %s\n", i, landmark.Name, landmark.Cost, landmark.Description)
	}

	fmt.Print("Which landmark do you want to buy? ")
	landmarkIdx, err := scanInt(choices)
	if err != nil {
		fmt.Println("No landmark selected.")
		return false
	}
	landmarkName := choiceNames[landmarkIdx-1]
	landmark := landmarkCards[landmarkName]

	if rlr.Coins.Total() < landmark.Cost {
		fmt.Printf("Player %d does not have enough coins to buy %s\n", rlr.ID, landmarkName)
	} else {
		fmt.Printf("Player %d buys %s\n", rlr.ID, landmarkName)
		rlr.Coins.TransferTo(landmark.Cost, &bank)
		rlr.LandmarkCards[landmarkName] = true
	}

	return true
}

func promptRoll(choice bool) (int, bool, error) {
	var dieCount int
	var err error
	var doubles bool

	if choice {
		fmt.Print("Roll 1 die or 2 dice? ")
		dieCount, err = scanInt([]int{1, 2})

		if err != nil {
			return 0, false, err
		}
	} else {
		dieCount = 1
	}

	r := 0
	for i := 0; i < dieCount; i++ {
		die := rand.Intn(5) + 1
		if i == 2 && r == die {
			doubles = true
		}
		r += die
	}

	// return r, doubles, nil
	return 3, doubles, nil
}

func printPlayerStatus() {
	for _, p := range plrs {
		fmt.Printf("%d: %d Coins\n", p.ID, p.Coins.Total())
		fmt.Println(p.SupplyCards)
	}
}
