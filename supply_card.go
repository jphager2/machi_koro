package main

type supplyCardCollection struct {
	Cards            []*supplyCard
	PrioritizedCards []*supplyCard
}

func newSupplyCardCollection(cards []*supplyCard) supplyCardCollection {
	var prioritized []*supplyCard

	for i := 0; i < 3; i++ {
		for _, card := range cards {
			if card.Effect.Priority == i {
				prioritized = append(prioritized, card)
			}
		}
	}

	return supplyCardCollection{
		Cards:            cards,
		PrioritizedCards: prioritized,
	}
}

func (s *supplyCardCollection) FindByIcon(icon string) []*supplyCard {
	var found []*supplyCard

	for i := range s.Cards {
		card := s.Cards[i]

		if card.Icon == icon {
			found = append(found, card)
		}
	}

	return found
}

func (s *supplyCardCollection) FindByName(name string) *supplyCard {
	for i := range s.Cards {
		card := s.Cards[i]

		if name == card.Name {
			return card
		}
	}

	return &supplyCard{}
}

func (s *supplyCardCollection) FindByRoll(roll int) []*supplyCard {
	var found []*supplyCard

	for i := range s.PrioritizedCards {
		card := s.Cards[i]

		for _, number := range card.ActiveNumbers {
			if number == roll {
				found = append(found, card)
				break
			}
		}
	}

	return found
}

type supplyCard struct {
	Name          string
	Cost          int
	ActiveNumbers []int
	Effect        effect
	Icon          string
	Supply        int
}
