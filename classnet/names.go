package classnet

import "math/rand"

// When clients join a room they are assigned a name.
// Names are a color + an animal.

// Colors is a list of colors.
var Colors = []string{
	"red",
	"orange",
	"yellow",
	"green",
	"blue",
	"purple",
	"pink",
	"brown",
}

// Animals is a list of animals (1 per letter of the alphabet).
var Animals = []string{
	"alligator",
	"bear",
	"cat",
	"dog",
	"elephant",
	"frog",
	"giraffe",
	"hippo",
	"iguana",
	"jaguar",
	"kangaroo",
	"lion",
	"monkey",
	"newt",
	"octopus",
	"penguin",
	"quail",
	"rabbit",
	"snake",
	"tiger",
	"unicorn",
	"vulture",
	"walrus",
	"x-ray",
	"yak",
	"zebra",
}

// Name is a color + an animal.
type Name struct {
	Color  string `json:"color"`
	Animal string `json:"animal"`
}

// NewName creates a new name.
func NewName() Name {
	color := Colors[rand.Intn(len(Colors))]
	animal := Animals[rand.Intn(len(Animals))]
	return Name{
		Color:  color,
		Animal: animal,
	}
}
