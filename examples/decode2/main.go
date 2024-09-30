package main

import (
	"encoding/hex"
	"fmt"

	"github.com/defiweb/go-rlp"
)

func main() {
	data, _ := hex.DecodeString("c983666f6f836261722a")

	// Decode the data
	dec, _, err := rlp.DecodeLazy(data)
	if err != nil {
		panic(err)
	}

	// Check if the decoded data is a list
	list, err := dec.List()
	if err != nil {
		panic(err)
	}
	if len(list) != 3 {
		panic("invalid list length")
	}

	// Check if list items are strings
	if !list[0].IsString() || !list[1].IsString() || !list[2].IsString() {
		panic("unexpected types")
	}

	// Decode items
	foo, _ := list[0].String()
	bar, _ := list[1].String()
	num, _ := list[2].Uint()

	// Print the decoded data
	fmt.Println(foo.Get())
	fmt.Println(bar.Get())
	fmt.Println(num.Get())
}
