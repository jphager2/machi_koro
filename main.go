package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Player struct {
	Id            int
	SupplyCards   map[string]int
	LandmarkCards map[string]bool
	Coins         CoinSet
}

func GetSupplyCardPurchase(roller *Player) bool {
	var res int

	fmt.Printf("Do you want to buy an establishment? (%d coins) ", roller.Coins.Total())
	res, _ = GetInt([]int{0, 1})

	if res != 1 {
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
	supplyCardIdx, err := GetInt(choices)
	if err != nil {
		fmt.Println("No establishment selected.")
		return false
	}
	supplyCardName := choiceNames[supplyCardIdx-1]
	card := supplyCards.FindByName(supplyCardName)

	if roller.Coins.Total() < card.Cost {
		fmt.Printf("Player %d does not have enough coins to buy %s\n", roller.Id, supplyCardName)
	} else {
		fmt.Printf("Player %d buys %s\n", roller.Id, supplyCardName)
		roller.Coins.WithdrawTo(card.Cost, &bank)
		roller.SupplyCards[supplyCardName]++
		card.Supply--
	}
	return true
}

func GetLandmarkCardPurchase(roller *Player) bool {
	var res int

	fmt.Printf("Do you want to buy a landmark? (%d coins) ", roller.Coins.Total())
	res, _ = GetInt([]int{0, 1})

	if res != 1 {
		return false
	}

	i := 0
	choices := []int{}
	choiceNames := []string{}
	fmt.Println("Landmarks: ")
	for _, landmark := range landmarkCardsSorted {
		if roller.LandmarkCards[landmark.Name] {
			continue
		}

		i++
		choices = append(choices, i)
		choiceNames = append(choiceNames, landmark.Name)
		fmt.Printf("  (%d) %s [%d coins]: %s\n", i, landmark.Name, landmark.Cost, landmark.Description)
	}

	fmt.Print("Which landmark do you want to buy? ")
	landmarkIdx, err := GetInt(choices)
	if err != nil {
		fmt.Println("No landmark selected.")
		return false
	}
	landmarkName := choiceNames[landmarkIdx-1]
	landmark := landmarkCards[landmarkName]

	if roller.Coins.Total() < landmark.Cost {
		fmt.Printf("Player %d does not have enough coins to buy %s\n", roller.Id, landmarkName)
	} else {
		fmt.Printf("Player %d buys %s\n", roller.Id, landmarkName)
		roller.Coins.WithdrawTo(landmark.Cost, &bank)
		roller.LandmarkCards[landmarkName] = true
	}

	return true
}

func GetRoll(roller *Player) (int, bool, error) {
	var dieCount int
	var err error
	var doubles bool

	if roller.LandmarkCards["Train Station"] {
		fmt.Print("Roll 1 or 2 dice? ")
		dieCount, err = GetInt([]int{1, 2})

		if err != nil {
			return 0, false, err
		}
	} else {
		dieCount = 1
	}

	roll := 0
	for i := 0; i < dieCount; i++ {
		die := rand.Intn(5) + 1
		if i == 2 && roll == die {
			doubles = true
		}
		roll += die
	}

	return roll, doubles, nil
}

func Roll(roller *Player, r int) {
	cards := supplyCards.FindByRoll(r)

	for _, card := range cards {
		card.Effect.Call(*card, roller, players)
	}
}

func PrintPlayerStatus() {
	for _, player := range players {
		fmt.Printf("%d: %d Coins\n", player.Id, player.Coins.Total())
		fmt.Println(player.SupplyCards)
	}
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
				fmt.Printf("Player %d gets %d coins from player %d [%s]\n", roller.Id, totalPayout, player.Id, card.Name)
				remainder := player.Coins.WithdrawTo(totalPayout, &roller.Coins)

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
				fmt.Printf("Player (%d) has %d coins\n", player.Id, player.Coins.Total())
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

				fmt.Printf("Player %d gets %d coins from player %d [%s]\n", roller.Id, totalPayout, player.Id, card.Name)
				remainder := player.Coins.WithdrawTo(totalPayout, &roller.Coins)

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
						fmt.Printf("Player (%d) has cards:\n", player.Id)
					}

					j := 1
					for cardName, cardCount := range player.SupplyCards {
						currentCard := supplyCards.FindByName(cardName)

						if currentCard.Icon == "Major" || cardCount == 0 {
							continue
						}
						cardChoices[player.Id] = append(cardChoices[player.Id], j)
						cardChoiceNames[player.Id] = append(cardChoiceNames[player.Id], cardName)
						fmt.Printf("  (%d) %s [%d]\n", j, cardName, cardCount)
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

				fmt.Printf("Player %d trades %s for %s with player %d [%s]\n", roller.Id, giveCardName, takeCardName, playerId, card.Name)
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

	landmarkCardsSorted []LandmarkCard
	landmarkCards       map[string]LandmarkCard

	players []*Player
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	supplyCards = SupplyCardCollection{
		Cards: []*SupplyCard{
			&SupplyCard{
				Name:          "Wheat Field",
				Cost:          1,
				ActiveNumbers: []int{1},
				Effect:        NewAllBankPayout(1),
				Icon:          "Wheat",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Ranch",
				Cost:          1,
				ActiveNumbers: []int{2},
				Effect:        NewAllBankPayout(1),
				Icon:          "Cow",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Bakery",
				Cost:          1,
				ActiveNumbers: []int{2, 3},
				Effect:        NewRollerBankPayout(1),
				Icon:          "Bread",
			},
			&SupplyCard{
				Name:          "Cafe",
				Cost:          2,
				ActiveNumbers: []int{3},
				Effect:        NewRollerPayout(1),
				Icon:          "Cup",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Convenience Store",
				Cost:          2,
				ActiveNumbers: []int{4},
				Effect:        NewRollerBankPayout(3),
				Icon:          "Bread",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Forest",
				Cost:          3,
				ActiveNumbers: []int{5},
				Effect:        NewAllBankPayout(1),
				Icon:          "Gear",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Stadium",
				Cost:          6,
				ActiveNumbers: []int{6},
				Effect:        stadiumEffect,
				Icon:          "Major",
				Supply:        4,
			},
			&SupplyCard{
				Name:          "TV Station",
				Cost:          7,
				ActiveNumbers: []int{6},
				Effect:        tvStationEffect,
				Icon:          "Major",
				Supply:        4,
			},
			&SupplyCard{
				Name:          "Business Center",
				Cost:          8,
				ActiveNumbers: []int{6},
				Effect:        businessCenterEffect,
				Icon:          "Major",
				Supply:        4,
			},
			&SupplyCard{
				Name:          "Cheese Factory",
				Cost:          5,
				ActiveNumbers: []int{7},
				Effect:        NewIconCardPayout(3, "Cow"),
				Icon:          "Factory",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Furniture Factory",
				Cost:          3,
				ActiveNumbers: []int{8},
				Effect:        NewIconCardPayout(3, "Gear"),
				Icon:          "Factory",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Mine",
				Cost:          6,
				ActiveNumbers: []int{9},
				Effect:        NewAllBankPayout(5),
				Icon:          "Gear",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Family Restaurant",
				Cost:          3,
				ActiveNumbers: []int{9, 10},
				Effect:        NewRollerPayout(2),
				Icon:          "Cup",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Apple Orchard",
				Cost:          3,
				ActiveNumbers: []int{10},
				Effect:        NewAllBankPayout(3),
				Icon:          "Wheat",
				Supply:        6,
			},
			&SupplyCard{
				Name:          "Fruit and Vegetable Market",
				Cost:          2,
				ActiveNumbers: []int{11, 12},
				Effect:        NewIconCardPayout(2, "Wheat"),
				Icon:          "Fruit",
				Supply:        6,
			},
		},
	}

	landmarkCardsSorted = []LandmarkCard{
		LandmarkCard{
			Name:        "Train Station",
			Cost:        4,
			Description: "You may roll 1 or 2 dice",
		},
		LandmarkCard{
			Name:        "Shopping Mall",
			Cost:        10,
			Description: "Each of your [Cup] and [Bread] establishments earn +1 coin",
		},
		LandmarkCard{
			Name:        "Amusement Park",
			Cost:        16,
			Description: "If you roll doubles take another turn after this one",
		},
		LandmarkCard{
			Name:        "Radio Tower",
			Cost:        22,
			Description: "Once every turn you can choose to re-roll your dice",
		},
	}
	landmarkCards = make(map[string]LandmarkCard)
	for _, card := range landmarkCardsSorted {
		landmarkCards[card.Name] = card
	}
}

func main() {
	fmt.Println("machi koro!")

	fmt.Print("How many players (2 - 4): ")
	playerCount, err := GetInt([]int{2, 3, 4})

	if err != nil {
		fmt.Println("This game is for 2-4 players")
		os.Exit(1)
	}

	for i := 0; i < playerCount; i++ {
		player := Player{Id: i}
		players = append(players, &player)
		remainder := bank.WithdrawTo(3, &player.Coins)

		if remainder > 0 {
			fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
		}
		player.SupplyCards = make(map[string]int)
		player.SupplyCards["Wheat Field"] += 1
		player.SupplyCards["Bakery"] += 1

		player.LandmarkCards = make(map[string]bool)
		for landmarkName, _ := range landmarkCards {
			player.LandmarkCards[landmarkName] = false
		}
	}

	turn := 0

	// Game Loop
	for {
		didAction := false
		roller := players[turn]

		fmt.Printf("It's player %d's turn\n", roller.Id)

		roll, doubles, err := GetRoll(roller)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Player %d rolls %d\n", roller.Id, roll)

		if roller.LandmarkCards["RadioTower"] {
			fmt.Print("Do you want to re-roll? ")
			reroll, _ := GetInt([]int{0, 1})

			if reroll == 1 {
				roll, _, err := GetRoll(roller)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Player %d rolls %d\n", roller.Id, roll)
			}
		}

		Roll(roller, roll)
		didAction = GetSupplyCardPurchase(roller)

		if didAction {
			didAction = false
		} else {
			didAction = GetLandmarkCardPurchase(roller)
		}

		winner := true
		for _, hasLandmark := range roller.LandmarkCards {
			if !hasLandmark {
				winner = false
				break
			}
		}
		if winner {
			fmt.Printf("Player %d has won the game!\n", roller.Id)
			os.Exit(0)
		}

		if doubles && roller.LandmarkCards["Amusement Park"] {
			continue
		}

		turn = (turn + 1) % len(players)
	}
}
