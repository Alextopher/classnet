package main

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
	Color  string
	Animal string
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

// String returns the name as a string.
func (name Name) String() string {
	return name.Color + " " + name.Animal
}

// MarshalText marshals the name as a JSON string.
func (name Name) MarshalText() ([]byte, error) {
	return []byte(`"` + name.String() + `"`), nil
}

// UnMarshalText unmarshals a JSON string to a name.
func (name *Name) UnmarshalText(b []byte) error {
	// Remove the quotes
	b = b[1 : len(b)-1]

	// Split the string into color and animal
	s := string(b)
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			name.Color = s[:i]
			name.Animal = s[i+1:]
			return nil
		}
	}

	return nil
}
