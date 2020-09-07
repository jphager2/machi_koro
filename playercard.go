package main

type playerCard struct {
	Total      int
	Renovation int
}

func (p *playerCard) Active() int {
	return p.Total - p.Renovation
}
