# go-rlp

The `go-rlp` package provides an implementation of RLP serialization format.

https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/

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
```

### Decoding data (method 1)

```go
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
```

### Decoding data (method 2)

This method does not require a prior definition of the data structure, making it useful when the data structure is not
known in advance.

```go
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
```

## Documentation

[https://pkg.go.dev/github.com/defiweb/go-rlp](https://pkg.go.dev/github.com/defiweb/go-rlp)
