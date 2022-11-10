# go-rlp

The `go-rlp` package provides an implementation of RLP serialization format.

## Installation

```bash
go get github.com/ethereum/go-rlp
```

## Usage

### Encoding data

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/defiweb/go-rlp"
)

func main() {
	list := rlp.NewList()
	list.Append(rlp.NewString("foo"))
	list.Append(rlp.NewString("bar"))
	list.Append(rlp.NewBigInt(big.NewInt(42)))

	enc, err := rlp.Encode(list)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x", enc)
}
```

### Decoding data (method 1)

```go
package main

import (
	"fmt"
	"encoding/hex"

	"github.com/defiweb/go-rlp"
)

func main() {
	data, _ := hex.DecodeString("c983666f6f836261722a")

	items := rlp.ListItem{
		&rlp.StringItem{},
		&rlp.StringItem{},
		&rlp.BigIntItem{},
	}

	// Decode the data
	_, err := rlp.DecodeInto(data, &items)
	if err != nil {
		panic(err)
	}

	// Print the decoded data
	fmt.Println(items[0].(*rlp, StringItem).String())
	fmt.Println(items[1].(*rlp, StringItem).String())
	fmt.Println(items[2].(*rlp, BigIntItem).X)
}
```

### Decoding data (method 2)

This method does not require a prior definition of the data structure, hence it can be useful when the data structure is
not known in advance.

```go
package main

import (
	"fmt"
	"encoding/hex"

	"github.com/defiweb/go-rlp"
)

func main() {
	data, _ := hex.DecodeString("c983666f6f836261722a")

	// Decode the data
	dec, _, err := rlp.Decode(data)
	if err != nil {
		panic(err)
	}

	// Check if the decoded data is a list
	list, err := dec.GetList()
	if err != nil {
		panic(err)
	}
	if len(list) != 3 {
		panic("invalid list length")
	}

	// Check if list items are strings
	if !list[0].IsString() || list[1].IsString() || list[2].IsString() {
		panic("expected strings")
	}

	// Decode items
	foo, _ := list[0].GetString()
	bar, _ := list[1].GetString()
	num, _ := list[2].GetBigInt()

	fmt.Println(foo, bar, num)
}
```

## Documentation

[https://pkg.go.dev/github.com/defiweb/go-rlp](https://pkg.go.dev/github.com/defiweb/go-rlp)
