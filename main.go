package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("machi koro!")

	fmt.Print("How many players (2 - 4): ")
	plrCount, err := scanInt([]int{2, 3, 4})

	if err != nil {
		fmt.Println("This game is for 2-4 players")
		os.Exit(1)
	}

	version, err := promptVersionChoice()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	version.Init()

	for i := 0; i < plrCount; i++ {
		p := player{ID: i}
		plrs = append(plrs, &p)
		remainder := bank.TransferTo(3, &p.Coins)

		if remainder > 0 {
			fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
		}
		p.SupplyCards = make(map[string]*playerCard)
		p.SupplyCards["Wheat Field"] = &playerCard{Total: 1, Renovation: 0}
		p.SupplyCards["Bakery"] = &playerCard{Total: 1, Renovation: 0}

		p.LandmarkCards = make(map[string]bool)
		for _, landmark := range landmarkCardsSorted {
			p.LandmarkCards[landmark.Name] = landmark.Cost == 0
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

		if r >= 10 && rlr.LandmarkCards["Harbor"] {
			fmt.Print("Do you want to add 2 to your roll? ")

			if res := promptBool(); res {
				r += 2
			}
		}

		// Card effects should be applied in priority order, first red cards, then
		// green/blue cards, then purple cards.
		cards := market.FindByRoll(r)
		// This two dice roll is used for some card effects to determine payouts.  It
		// should only be rolled once per roll.
		specialRoll := (rand.Intn(11) + 1)
		for i := 0; i < len(plrs); i++ {
			p := plrs[(len(plrs)+rlr.ID-i)%len(plrs)]

			for _, card := range cards {
				pc, ok := p.SupplyCards[card.Name]
				if !ok {
					continue
				}
				c := pc.Active()
				pc.Renovation = 0
				if c > 0 {
					card.Effect.Call(*card, rlr, p, c, pc, specialRoll)
				}
			}
		}

		if doubles && rlr.LandmarkCards["Amusement Park"] {
			fmt.Print("You got doubles, do you want to roll again? ")

			if res := promptBool(); res {
				continue
			}
		}

		if rlr.Coins.Total() == 0 && rlr.LandmarkCards["City Hall"] {
			fmt.Println("Getting 1 coin from the bank, since you didn't have any")

			remainder := bank.TransferTo(1, &rlr.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
		}

		didAction = promptSupplyCardPurchase(rlr)
		if didAction {
			didAction = false
		} else {
			didAction = promptLandmarkCardPurchase(rlr)
		}

		if !didAction && rlr.LandmarkCards["Airport"] {
			fmt.Println("Getting 10 coins from the bank, since you didn't buy anything")

			remainder := bank.TransferTo(10, &rlr.Coins)

			if remainder > 0 {
				fmt.Printf("Bank did not have enough money. Missing: %d\n", remainder)
			}
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

		turnInvestmentMax := rlr.SupplyCards["Tech Startup"].Total
		if turnInvestmentMax > 0 {
			promptInvestment(rlr, turnInvestmentMax)
		}

		turn = (turn + 1) % len(plrs)
	}
}

func promptInvestment(rlr *player, max int) {
	fmt.Printf("How much do you want to invest into your Tech Startups (max %d)\n", max)
	var choices []int
	for i := 0; i < max; i++ {
		choices = append(choices, i+1)
	}
	investment, err := scanInt(choices)
	if err != nil {
		fmt.Println("No investment made.")
		return
	}
	rlr.Coins.TransferTo(investment, &rlr.Investment)
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
	for _, cardCount := range market.EachCard() {
		card := cardCount.Card
		count := cardCount.Count
		if count == 0 {
			fmt.Printf("Count is 0!!! %v", card)
			continue
		}
		// Some cards have negative cost (i.e. get money from the bank)
		displayCost := card.Cost
		if displayCost < 0 {
			displayCost = 0
		}

		i++
		choices = append(choices, i)
		choiceNames = append(choiceNames, card.Name)
		fmt.Printf("  (%d) %s [%d coins] (%d left): %s\n", i, card.Name, displayCost, count, card.Effect.Description())
	}

	fmt.Print("Which establishment do you want to buy? ")
	supplyCardIdx, err := scanInt(choices)
	if err != nil {
		fmt.Println("No establishment selected.")
		return false
	}
	supplyCardName := choiceNames[supplyCardIdx-1]
	card := market.FindByName(supplyCardName)

	if rlr.Coins.Total() < card.Cost {
		fmt.Printf("Player %d does not have enough coins to buy %s\n", rlr.ID, supplyCardName)
	} else {
		fmt.Printf("Player %d buys %s\n", rlr.ID, supplyCardName)
		if card.Cost > 0 {
			rlr.Coins.TransferTo(card.Cost, &bank)
		} else if card.Cost < 0 {
			bank.TransferTo(card.Cost, &rlr.Coins)
		}
		rlr.SupplyCards[supplyCardName].Total++
		market.Purchase(card.Name)
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

func promptVersionChoice() (gameVersion, error) {
	var choice gameVersion
	choices := []int{}
	choiceNames := []string{}
	fmt.Println("Versions: ")
	for i, version := range gameVersionsSorted {
		choices = append(choices, i+1)
		choiceNames = append(choiceNames, version.Name)
		fmt.Printf("  (%d) %s \n", i+1, version.Name)
	}

	fmt.Print("Which version do you want to play? ")
	versionIdx, err := scanInt(choices)
	if err != nil {
		return choice, errors.New("No version selected.")
	}

	return gameVersionsSorted[versionIdx-1], nil
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
