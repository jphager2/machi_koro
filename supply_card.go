package main

type SupplyCardCollection struct {
	Cards []*SupplyCard
}

func (s *SupplyCardCollection) FindByIcon(icon string) []*SupplyCard {
	var found []*SupplyCard

	for i := range s.Cards {
		card := s.Cards[i]

		if card.Icon == icon {
			found = append(found, card)
		}
	}

	return found
}

func (s *SupplyCardCollection) FindByName(name string) *SupplyCard {
	for i := range s.Cards {
		card := s.Cards[i]

		if name == card.Name {
			return card
		}
	}

	return &SupplyCard{}
}

func (s *SupplyCardCollection) FindByRoll(roll int) []*SupplyCard {
	var found []*SupplyCard

	for i := range s.Cards {
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

type SupplyCard struct {
	Name          string
	Cost          int
	ActiveNumbers []int
	Effect        Effect
	Icon          string
	Supply        int
}
