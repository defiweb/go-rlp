package main

import (
	"encoding/hex"
	"fmt"

	"github.com/defiweb/go-rlp"
)

func main() {
	data, _ := hex.DecodeString("c983666f6f836261722a")

	// Define the data structure
	var item1 rlp.String
	var item2 rlp.String
	var item3 rlp.Uint
	list := rlp.List{&item1, &item2, &item3}

	// Decode the data
	_, err := rlp.Decode(data, &list)
	if err != nil {
		panic(err)
	}

	// Print the decoded data
	fmt.Println(item1.Get())
	fmt.Println(item2.Get())
	fmt.Println(item3.Get())
}
