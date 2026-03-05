package services

import (
	"fmt"
	"math/rand"
)

var anonAdjectives = []string{
	"Amber", "Ancient", "Arctic", "Atomic", "Blazing",
	"Bold", "Brave", "Bright", "Bronze", "Calm",
	"Cosmic", "Crimson", "Crystal", "Dark", "Daring",
	"Electric", "Ember", "Epic", "Fierce", "Frosty",
	"Gilded", "Gleaming", "Golden", "Grim", "Happy",
	"Hidden", "Iron", "Ivory", "Jade", "Jolly",
	"Keen", "Kind", "Lunar", "Misty", "Mystic",
	"Noble", "Neon", "Nimble", "Obsidian", "Onyx",
	"Primal", "Quick", "Radiant", "Rapid", "Ruby",
	"Rustic", "Sacred", "Scarlet", "Shadow", "Silver",
	"Silent", "Sleek", "Solar", "Sonic", "Steel",
	"Stormy", "Swift", "Tidal", "Turbo", "Twilight",
	"Ultra", "Velvet", "Vibrant", "Vivid", "Wild",
	"Windy", "Wise", "Zesty",
}

var anonNouns = []string{
	"Anchor", "Anvil", "Arrow", "Axe", "Bear",
	"Bolt", "Boulder", "Cedar", "Cinder", "Cliff",
	"Cloud", "Coal", "Comet", "Coral", "Crane",
	"Dagger", "Dawn", "Dune", "Eagle", "Ember",
	"Falcon", "Fang", "Flint", "Fog", "Forest",
	"Frost", "Gale", "Glacier", "Hawk", "Heron",
	"Inferno", "Jaguar", "Lava", "Lion", "Lynx",
	"Mesa", "Moon", "Nova", "Oak", "Opal",
	"Orca", "Peak", "Pepper", "Puma", "Quartz",
	"Raven", "Ridge", "River", "Rock", "Saber",
	"Sage", "Shark", "Shore", "Spark", "Spire",
	"Star", "Summit", "Tiger", "Timber", "Torch",
	"Viper", "Volt", "Wave", "Wolf", "Wood",
	"Zenith",
}

func randomAnonName() string {
	adj := anonAdjectives[rand.Intn(len(anonAdjectives))]
	noun := anonNouns[rand.Intn(len(anonNouns))]
	suffix := rand.Intn(10000)
	return fmt.Sprintf("%s%s%04d", adj, noun, suffix)
}
