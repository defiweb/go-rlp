package rlp

import (
	"math/big"
)

// RLP is a raw RLP encoded data that can be decoded into any other type later.
type RLP []byte

// Decode decodes RLP encoded data into the given value.
func (r RLP) Decode(dst Decoder) error {
	_, err := dst.DecodeRLP(r)
	return err
}

// Length returns the length of the string or number of items in the list.
// If the item is invalid, it returns 0.
func (r RLP) Length() int {
	switch {
	case r.IsString():
		_, dataLen, _, _ := decodePrefix(r)
		return int(dataLen)
	case r.IsList():
		_, dataLen, prefixLen, err := decodePrefix(r)
		if err != nil {
			return 0
		}
		totalLen := int(dataLen + uint64(prefixLen))
		if len(r) < totalLen {
			return 0
		}
		n := 0
		d := r[prefixLen:totalLen]
		for ; len(d) > 0; n++ {
			_, dataLen, prefixLen, err := decodePrefix(d)
			if err != nil {
				return 0
			}
			totalLen := int(dataLen + uint64(prefixLen))
			if totalLen > len(d) || totalLen == 0 {
				return 0
			}
			d = d[totalLen:]
		}
		return n
	}
	return 0
}

// String attempts to decode itself as a string. If the decoding is
// successful, it returns the decoded string.
func (r RLP) String() (v String, err error) {
	_, err = (&v).DecodeRLP(r)
	return
}

// Bytes attempts to decode itself as a byte slice. If the decoding is
// successful, it returns the decoded byte slice.
func (r RLP) Bytes() (v Bytes, err error) {
	_, err = (&v).DecodeRLP(r)
	return
}

// List attempts to decode itself as a list. If the decoding is
// successful, it returns the decoded list.
func (r RLP) List() (l TypedList[RLP], err error) {
	_, err = (&l).DecodeRLP(r)
	return
}

// Uint attempts to decode itself as an uint64. If the decoding is
// successful, it returns the decoded uint64.
func (r RLP) Uint() (v Uint, err error) {
	_, err = (&v).DecodeRLP(r)
	return
}

// BigInt attempts to decode itself as a big.Int. If the decoding is
// successful, it returns the decoded big.Int.
func (r RLP) BigInt() (v *BigInt, err error) {
	v = new(BigInt)
	_, err = v.DecodeRLP(r)
	return
}

// IsString returns true if the encoded data is an RLP string.
// Do not confuse this with the Go string type; an RLP string could be decoded
// as a string, byte slice, uint64, or big.Int.
// If the RLP data is empty, it returns false.
func (r RLP) IsString() bool {
	if len(r) == 0 {
		return false
	}
	return r[0] <= longStringMax
}

// IsList returns true if the encoded data is an RLP list.
// If the RLP data is empty, it returns false.
func (r RLP) IsList() bool {
	if len(r) == 0 {
		return false
	}
	return r[0] >= listOffset
}

// EncodeRLP implements the Encoder interface.
func (r RLP) EncodeRLP() ([]byte, error) {
	return r, nil
}

// DecodeRLP implements the Decoder interface.
func (r *RLP) DecodeRLP(data []byte) (int, error) {
	_, dataLen, prefixLen, err := decodePrefix(data)
	if err != nil {
		return 0, err
	}
	totalLen := int(dataLen + uint64(prefixLen))
	if totalLen > len(data) {
		return 0, ErrUnexpectedEndOfData
	}
	*r = data[:totalLen]
	return totalLen, nil
}

// String is a string type that can be encoded and decoded to/from RLP.
type String string

// Get returns the string value.
func (s String) Get() string {
	return string(s)
}

// Set sets the string value.
func (s *String) Set(value string) {
	*s = String(value)
}

// EncodeRLP implements the Encoder interface.
func (s String) EncodeRLP() ([]byte, error) {
	return encodeString(string(s))
}

// DecodeRLP implements the Decoder interface.
func (s *String) DecodeRLP(data []byte) (int, error) {
	return decodeString(data, (*string)(s))
}

// Bytes is a byte slice type that can be encoded and decoded to/from RLP.
type Bytes []byte

// Get returns the byte slice value.
func (b Bytes) Get() []byte {
	return b
}

// Set sets the byte slice value.
func (b *Bytes) Set(value []byte) {
	*b = value
}

// Ptr returns a pointer to the byte slice value.
func (b *Bytes) Ptr() *[]byte {
	return (*[]byte)(b)
}

// EncodeRLP implements the Encoder interface.
func (b Bytes) EncodeRLP() ([]byte, error) {
	return encodeBytes(b)
}

// DecodeRLP implements the Decoder interface.
func (b *Bytes) DecodeRLP(data []byte) (int, error) {
	return decodeBytes(data, (*[]byte)(b))
}

// Uint is an uint64 type that can be encoded and decoded to/from RLP.
type Uint uint64

// Get returns the uint64 value.
func (u Uint) Get() uint64 {
	return uint64(u)
}

// Ptr returns a pointer to the uint64 value.
func (u *Uint) Ptr() *uint64 {
	return (*uint64)(u)
}

// Set sets the uint64 value.
func (u *Uint) Set(value uint64) {
	*u = Uint(value)
}

// EncodeRLP implements the Encoder interface.
func (u Uint) EncodeRLP() ([]byte, error) {
	return encodeUint(uint64(u))
}

// DecodeRLP implements the Decoder interface.
func (u *Uint) DecodeRLP(data []byte) (int, error) {
	return decodeUint(data, (*uint64)(u))
}

// BigInt is a big.Int type that can be encoded and decoded to/from RLP.
type BigInt big.Int

// Get returns the big.Int value.
func (b BigInt) Get() *big.Int {
	return (*big.Int)(&b)
}

// Ptr returns a pointer to the big.Int value.
func (b *BigInt) Ptr() *big.Int {
	return (*big.Int)(b)
}

// Set sets the big.Int value.
func (b *BigInt) Set(value *big.Int) {
	(*big.Int)(b).Set(value)
}

// EncodeRLP implements the Encoder interface.
func (b BigInt) EncodeRLP() ([]byte, error) {
	return encodeBigInt((*big.Int)(&b))
}

// DecodeRLP implements the Decoder interface.
func (b *BigInt) DecodeRLP(data []byte) (int, error) {
	return decodeBigInt(data, (*big.Int)(b))
}

// List represents a list of RLP items.
//
// List items must implement the Encoder interface if the list is being encoded,
// or the Decoder interface if the list is being decoded. Otherwise, the encoding
// or decoding will fail.
//
// During decoding, the data is decoded into existing items if they are already
// in the list. Otherwise, the items are decoded into RLP types.
type List []any

// Get returns the slice of items.
func (l List) Get() []any {
	return l
}

// Ptr returns a pointer to the slice of items.
func (l *List) Ptr() *[]any {
	return (*[]any)(l)
}

// Set sets the slice of items.
func (l *List) Set(items ...any) {
	*l = items
}

// Add appends the given items to the list.
func (l *List) Add(items ...any) {
	*l = append(*l, items...)
}

// EncodeRLP implements the Encoder interface.
func (l List) EncodeRLP() ([]byte, error) {
	return encodeList(l)
}

// DecodeRLP implements the Decoder interface.
func (l *List) DecodeRLP(data []byte) (int, error) {
	return decodeList(data, (*[]any)(l))
}

// TypedList represents a RLP list of a specific type.
//
// The T type must implement the Encoder interface if the list is being encoded,
// or the Decoder interface if the list is being decoded. Otherwise, the encoding
// or decoding will fail.
//
// During decoding, the data is decoded into existing items if they are already
// in the list. If the list is shorter than the data, new items are appended to
// the list.
type TypedList[T any] []*T

// Get returns the slice of items.
func (l TypedList[T]) Get() []*T {
	return l
}

// Ptr returns a pointer to the slice of items.
func (l *TypedList[T]) Ptr() *[]*T {
	return (*[]*T)(l)
}

// Set sets the slice of items.
func (l *TypedList[T]) Set(items ...*T) {
	*l = items
}

// Add appends the given items to the list.
func (l *TypedList[T]) Add(items ...*T) {
	*l = append(*l, items...)
}

// EncodeRLP implements the Encoder interface.
func (l TypedList[T]) EncodeRLP() ([]byte, error) {
	return encodeTypedList(l)
}

// DecodeRLP implements the Decoder interface.
func (l *TypedList[T]) DecodeRLP(data []byte) (int, error) {
	return decodeTypedList(data, (*[]*T)(l), func() *T { return new(T) })
}
