package main

// TODO: Marketplace should manage the available and supply of cards
//       For example in harbor and millionaire's row not all cards are on the table at all times
//       In the english version it simply picks 10 cards at time. In the Czech version there are
//         * 5 cards 1-6
//         * 5 cards 7+
//         * 2 cards Major establishments
//
type marketplace struct {
	Cards            []*supplyCard
	PrioritizedCards []*supplyCard
}

func newMarketplace(cards []*supplyCard) marketplace {
	var prioritized []*supplyCard

	for i := 0; i < 3; i++ {
		for _, card := range cards {
			if card.Effect.Priority == i {
				prioritized = append(prioritized, card)
			}
		}
	}

	return marketplace{
		Cards:            cards,
		PrioritizedCards: prioritized,
	}
}

func (s *marketplace) FindByIcon(icon string) []*supplyCard {
	var found []*supplyCard

	for i := range s.Cards {
		card := s.Cards[i]

		if card.Icon == icon {
			found = append(found, card)
		}
	}

	return found
}

func (s *marketplace) FindByName(name string) *supplyCard {
	for i := range s.Cards {
		card := s.Cards[i]

		if name == card.Name {
			return card
		}
	}

	return &supplyCard{}
}

func (s *marketplace) FindByRoll(roll int) []*supplyCard {
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
