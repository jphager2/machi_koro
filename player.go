package main

type player struct {
	ID            int
	SupplyCards   map[string]*playerCard
	LandmarkCards map[string]bool
	Coins         coinSet
	Investment    coinSet
}
