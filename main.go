package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type player struct {
	ID            int
	SupplyCards   map[string]int
	LandmarkCards map[string]bool
	Coins         coinSet
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

		dieCount, err := promptDieCount(rlr.LandmarkCards["Train Station"] /* choice */)
		if err != nil {
			fmt.Println(err)
			continue
		}
		r, doubles := roll(dieCount)
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

func promptDieCount(choice bool) (int, error) {
	var dieCount int
	var err error

	if choice {
		fmt.Print("Roll 1 die or 2 dice? ")
		dieCount, err = scanInt([]int{1, 2})

		if err != nil {
			return 0, err
		}
	} else {
		dieCount = 1
	}

	return dieCount, err
}

func roll(dieCount int) (int, bool) {
	var doubles bool

	r := 0
	for i := 0; i < dieCount; i++ {
		die := rand.Intn(5) + 1
		if i == 2 && r == die {
			doubles = true
		}
		r += die
	}

	return r, doubles
}

func printPlayerStatus() {
	for _, p := range plrs {
		fmt.Printf("%d: %d Coins\n", p.ID, p.Coins.Total())
		fmt.Println(p.SupplyCards)
	}
}
