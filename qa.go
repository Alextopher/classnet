package main

import "math/rand"

var symbols = "0123456789ABCDEF"

// Question-Answer table
type QATable map[string]string

// Table are made up of "symbols"
func randomSymbol() string {
	symbol := make([]byte, 4)
	for i := range symbol {
		symbol[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(symbol)
}

func NewQATable() QATable {
	table := make(map[string]string)

	// Generate 16 random symbols
	for i := 0; i < 16; i++ {
		// Generate a symbol that isn't already in the table
		var symbol string
		for {
			symbol = randomSymbol()
			if _, ok := table[symbol]; !ok {
				break
			}
		}

		// Generate a random question
		question := randomSymbol()

		// Add the symbol and question to the table
		table[symbol] = question
	}

	return table
}
