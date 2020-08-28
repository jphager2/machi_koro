package main

import (
	"errors"
	"fmt"
	"math/rand"
)

// TODO: Marketplace should manage the available and supply of cards
//       For example in harbor and millionaire's row not all cards are on the table at all times
//       In the english version it simply picks 10 cards at time. In the Czech version there are
//         * 5 cards 1-6
//         * 5 cards 7+
//         * 2 cards Major establishments
//
type marketplace struct {
	Market           marketManager
	Cards            []*supplyCard
	PrioritizedCards []*supplyCard
}

type cardCount struct {
	Count int
	Card  *supplyCard
}

type marketManager interface {
	remove(string) error
	cards() []cardCount
}

type basicMarket struct {
	OnMarket []*supplyCard
}

type expansionMarket struct {
	LOnMarket map[string]int
	HOnMarket map[string]int
	MOnMarket map[string]int
	LCards    []*supplyCard
	HCards    []*supplyCard
	MCards    []*supplyCard
	Cards     []*supplyCard
}

func newMarketplace(cards []*supplyCard, manager marketManager) marketplace {
	var prioritized []*supplyCard

	for i := 0; i < 3; i++ {
		for _, card := range cards {
			if card.Effect.Priority == i {
				prioritized = append(prioritized, card)
			}
		}
	}

	return marketplace{
		Market:           manager,
		Cards:            cards,
		PrioritizedCards: prioritized,
	}
}

func newBasicMarketManager(cards []*supplyCard) basicMarket {
	return basicMarket{OnMarket: cards}
}

func newExpansionMarketManager(cards []*supplyCard) expansionMarket {
	var lcards []*supplyCard
	var hcards []*supplyCard
	var mcards []*supplyCard

	for _, card := range cards {
		if card.Icon == "Major" {
			mcards = append(mcards, card)
		} else if card.ActiveNumbers[0] >= 7 {
			hcards = append(hcards, card)
		} else {
			lcards = append(lcards, card)
		}
	}

	manager := expansionMarket{
		LCards:    lcards,
		HCards:    hcards,
		MCards:    mcards,
		LOnMarket: make(map[string]int),
		HOnMarket: make(map[string]int),
		MOnMarket: make(map[string]int),
		Cards:     cards,
	}

	manager.resupplyMarket()

	fmt.Printf("%v", manager.LOnMarket)
	fmt.Println(manager.HOnMarket)
	fmt.Println(manager)

	return manager
}

func newBasicMarketplace(cards []*supplyCard) marketplace {
	manager := newBasicMarketManager(cards)
	return newMarketplace(cards, manager)
}

func newExpansionMarketplace(cards []*supplyCard) marketplace {
	manager := newExpansionMarketManager(cards)
	return newMarketplace(cards, manager)
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

func findByName(cards []*supplyCard, name string) (*supplyCard, bool) {
	for _, card := range cards {
		if name == card.Name {
			return card, true
		}
	}

	return &supplyCard{}, false
}

func (s *marketplace) FindByName(name string) *supplyCard {
	card, _ := findByName(s.Cards, name)
	return card
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

func (s *marketplace) EachCard() []cardCount {
	return s.Market.cards()
}

func (s *marketplace) Purchase(name string) error {
	return s.Market.remove(name)
}

func (m basicMarket) remove(name string) error {
	card, _ := findByName(m.OnMarket, name)

	if card.Supply == 0 {
		return errors.New("Can't remove card, there are no more left")
	}

	card.Supply--

	return nil
}

func (m basicMarket) cards() []cardCount {
	var cards []cardCount
	for _, card := range m.OnMarket {
		cards = append(cards, cardCount{Count: card.Supply, Card: card})
	}
	return cards
}

func (m expansionMarket) remove(name string) error {
	var deleted bool
	var ok bool

	m.LOnMarket, deleted, ok = removePart(m.LOnMarket, name)
	if deleted {
		m.resupplyMarket()
	}
	if ok {
		return nil
	}

	m.HOnMarket, deleted, ok = removePart(m.HOnMarket, name)
	if deleted {
		m.resupplyMarket()
	}
	if ok {
		return nil
	}

	m.MOnMarket, deleted, ok = removePart(m.MOnMarket, name)
	if deleted {
		m.resupplyMarket()
	}
	if ok {
		return nil
	}

	return errors.New("Can't remove card, there are none on the market place")
}

func (m expansionMarket) cards() []cardCount {
	var cards []cardCount
	for name, count := range m.LOnMarket {
		card, _ := findByName(m.Cards, name)
		cards = append(cards, cardCount{Count: count, Card: card})
	}
	for name, count := range m.HOnMarket {
		card, _ := findByName(m.Cards, name)
		cards = append(cards, cardCount{Count: count, Card: card})
	}
	for name, count := range m.MOnMarket {
		card, _ := findByName(m.Cards, name)
		cards = append(cards, cardCount{Count: count, Card: card})
	}
	return cards
}

func takeCard(cards []*supplyCard) (*supplyCard, int) {
	idx := rand.Intn(len(cards)-1) + 1
	return cards[idx], idx
}

func removePart(count map[string]int, name string) (map[string]int, bool, bool) {
	var deleted bool

	c, ok := count[name]
	if !ok {
		return count, deleted, ok
	}

	c--

	if c > 0 {
		count[name] = c
	} else {
		delete(count, name)
		deleted = true
	}

	return count, deleted, ok
}

func resupplyPart(count map[string]int, cards []*supplyCard, max int) (map[string]int, []*supplyCard) {
	for len(count) < max && len(cards) > 0 {
		card, idx := takeCard(cards)
		if c, ok := count[card.Name]; ok {
			count[card.Name] = c + 1
		} else {
			count[card.Name] = 1
		}
		card.Supply--
		if card.Supply == 0 {
			cards = append(cards[0:idx], cards[idx+1:len(cards)]...)
		}
	}

	return count, cards
}

func (m expansionMarket) resupplyMarket() {
	m.LOnMarket, m.LCards = resupplyPart(m.LOnMarket, m.LCards, 5)
	m.HOnMarket, m.HCards = resupplyPart(m.HOnMarket, m.HCards, 5)
	m.MOnMarket, m.MCards = resupplyPart(m.MOnMarket, m.MCards, 2)
}
