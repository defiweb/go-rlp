package main

import (
	"fmt"

	"github.com/defiweb/go-rlp"
)

func main() {
	list := rlp.List{
		rlp.String("foo"),
		rlp.String("bar"),
		rlp.Uint(42),
	}

	enc, err := rlp.Encode(list)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x", enc)
}
