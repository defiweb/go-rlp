package rlp

import (
	"bytes"
	"math/big"
)

// NewString creates a new String item.
func NewString(data string) *StringItem {
	return NewBytes([]byte(data))
}

// NewBytes creates a new String item from byte slice.
func NewBytes(data []byte) *StringItem {
	return (*StringItem)(&data)
}

// NewList creates a new List item.
func NewList(items ...Item) *ListItem {
	return (*ListItem)(&items)
}

// NewUint creates a new Uint item.
func NewUint(x uint64) *UintItem {
	return &UintItem{X: x}
}

// NewBigInt creates a new BigInt item.
func NewBigInt(x *big.Int) *BigIntItem {
	return &BigIntItem{X: x}
}

// RLP is a raw RLP encoded data that can be decoded into any other type later.
type RLP []byte

// Bytes returns raw RLP encoded data. To get the decoded data use
// one of the Get* methods.
func (r RLP) Bytes() []byte {
	return r
}

// DecodeTo decodes RLP encoded data into the given item.
func (r RLP) DecodeTo(dest Item) error {
	_, err := dest.DecodeRLP(r)
	return err
}

// Length returns the length of the string or list. If item is invalid
// it returns 0.
func (r RLP) Length() uint64 {
	_, dataLen, _, _ := decodePrefix(r)
	return dataLen
}

// GetStringItem tries to decode itself as a string. If the decoding was
// successful it returns the decoded StringItem.
func (r RLP) GetStringItem() (*StringItem, error) {
	s := StringItem{}
	if err := r.DecodeTo(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetListItem tries to decode itself as a list. If the decoding was
// successful it returns the decoded ListItem.
func (r RLP) GetListItem() (*ListItem, error) {
	i := ListItem{}
	if err := r.DecodeTo(&i); err != nil {
		return nil, err
	}
	return &i, nil
}

// GetUintItem tries to decode itself as an integer. If the decoding was
// successful it returns the decoded UintItem.
func (r RLP) GetUintItem() (*UintItem, error) {
	i := UintItem{}
	if err := r.DecodeTo(&i); err != nil {
		return nil, err
	}
	return &i, nil
}

// GetBigIntItem tries to decode itself as a big integer. If the decoding was
// successful it returns the decoded BigIntItem.
func (r RLP) GetBigIntItem() (*BigIntItem, error) {
	i := BigIntItem{}
	if err := r.DecodeTo(&i); err != nil {
		return nil, err
	}
	return &i, nil
}

// Get decodes RLP encoded data into the given item and if decoding was
// successful it invokes the fn callback.
func (r RLP) Get(i Item, fn func(Item)) error {
	if err := r.DecodeTo(i); err != nil {
		return err
	}
	fn(i)
	return nil
}

// GetString tries to decode itself as a string. If the decoding was
// successful it returns the decoded string.
func (r RLP) GetString() (string, error) {
	s, err := r.GetStringItem()
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

// GetBytes tries to decode itself as a byte slice. If the decoding was
// successful it returns the decoded byte slice.
func (r RLP) GetBytes() ([]byte, error) {
	i, err := r.GetStringItem()
	if err != nil {
		return nil, err
	}
	return i.Bytes(), nil
}

// GetList tries to decode itself as a slice of RLP items. If the decoding was
// successful it returns the decoded slice.
func (r RLP) GetList() ([]*RLP, error) {
	i, err := r.GetListItem()
	if err != nil {
		return nil, err
	}
	var items []*RLP
	for _, v := range i.Items() {
		items = append(items, v.(*RLP))
	}
	return items, nil
}

// GetUint tries to decode itself as an uint64. If the decoding was
// successful it returns the decoded uint64.
func (r RLP) GetUint() (uint64, error) {
	i, err := r.GetUintItem()
	if err != nil {
		return 0, err
	}
	return i.X, nil
}

// GetBigInt tries to decode itself as a big.Int. If the decoding was
// successful it returns the decoded big.Int.
func (r RLP) GetBigInt() (*big.Int, error) {
	i, err := r.GetBigIntItem()
	if err != nil {
		return nil, err
	}
	return i.X, nil
}

// IsString returns true if the encoded data is a string.
// If the data is empty, it returns false.
func (r RLP) IsString() bool {
	if len(r) == 0 {
		return false
	}
	return r[0] <= longStringMax
}

// IsList returns true if the encoded data is a list.
// If the data is empty, it returns false.
func (r RLP) IsList() bool {
	if len(r) == 0 {
		return false
	}
	return r[0] >= listOffset
}

func (r RLP) EncodeRLP() ([]byte, error) {
	return r, nil
}

func (r *RLP) DecodeRLP(data []byte) (int, error) {
	_, dataLen, prefixLen, err := decodePrefix(data)
	if err != nil {
		return 0, err
	}
	if len(data) < int(dataLen)+prefixLen {
		return 0, ErrUnexpectedEndOfData
	}
	*r = data[:dataLen+uint64(prefixLen)]
	return int(dataLen + uint64(prefixLen)), nil
}

// StringItem is a RLP encoded string.
type StringItem []byte

// String returns the string representation of the string type.
func (s StringItem) String() string {
	return string(s)
}

// Bytes returns the byte slice representation of the string type.
func (s StringItem) Bytes() []byte {
	return s
}

func (s StringItem) EncodeRLP() ([]byte, error) {
	if len(s) == 0 {
		// Empty string.
		return []byte{stringOffset}, nil
	}
	if len(s) == 1 && s[0] < stringOffset {
		// Single byte string in the range [0x00, 0x7F].
		return []byte{s[0]}, nil
	}
	prefix, err := encodePrefix(uint64(len(s)), stringOffset)
	if err != nil {
		return nil, err
	}
	return append(prefix, s...), nil
}

func (s *StringItem) DecodeRLP(data []byte) (int, error) {
	if len(data) == 0 {
		// The data should not be empty. An empty string is encoded as a single
		// byte 0x80.
		return 0, ErrUnexpectedEndOfData
	}
	if data[0] == stringOffset {
		// The data is an empty string.
		*s = []byte{}
		return 1, nil
	}
	if data[0] < stringOffset {
		// Single byte in the range [0x00, 0x7F].
		*s = data[:1]
		return 1, nil
	}
	offset, dataLen, prefixLen, err := decodePrefix(data)
	if err != nil {
		return 0, err
	}
	if offset != stringOffset {
		return 0, ErrUnsupportedType
	}
	if uint64(len(data)) < dataLen+uint64(prefixLen) {
		return 0, ErrUnexpectedEndOfData
	}
	*s = data[prefixLen : dataLen+uint64(prefixLen)]
	return int(dataLen + uint64(prefixLen)), nil
}

// ListItem is a RLP encoded list. During decoding, the data is decoded
// into existing items if possible. If there are more items in the list
// than existing items, new items are created.
type ListItem []Item

// Items returns a slice of items in the list.
func (l *ListItem) Items() []Item {
	return *l
}

// Append appends an item to the list.
func (l *ListItem) Append(items ...Item) {
	*l = append(*l, items...)
}

func (l ListItem) EncodeRLP() ([]byte, error) {
	var buf bytes.Buffer
	for _, item := range l {
		data, err := item.EncodeRLP()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}
	prefix, err := encodePrefix(uint64(buf.Len()), listOffset)
	if err != nil {
		return nil, err
	}
	return append(prefix, buf.Bytes()...), nil
}

func (l *ListItem) DecodeRLP(data []byte) (int, error) {
	if len(data) == 0 {
		// The data should not be empty.
		return 0, ErrUnexpectedEndOfData
	}
	if data[0] == listOffset {
		// The data is an empty list.
		*l = []Item{}
		return 1, nil
	}
	offset, dataLen, prefixLen, err := decodePrefix(data)
	if err != nil {
		return 0, err
	}
	if offset != listOffset {
		return 0, ErrUnsupportedType
	}
	if uint64(len(data)) < dataLen+uint64(prefixLen) {
		return 0, ErrUnexpectedEndOfData
	}
	data = data[prefixLen : dataLen+uint64(prefixLen)]
	for n := 0; len(data) > 0; n++ {
		if n < len(*l) {
			// Decode data into existing item.
			itemLen, err := (*l)[n].DecodeRLP(data)
			if err != nil {
				return 0, err
			}
			data = data[itemLen:]
		} else {
			// Decode data into a new item. The data is decoded into a RLP item,
			// so it will be possible to decode it into a more specific type
			// later.
			item := &RLP{}
			itemLen, err := item.DecodeRLP(data)
			if err != nil {
				return 0, err
			}
			*l = append(*l, item)
			data = data[itemLen:]
		}
	}
	return int(dataLen + uint64(prefixLen)), nil
}

// UintItem is a RLP encoded unsigned integer.
type UintItem struct{ X uint64 }

func (u UintItem) EncodeRLP() ([]byte, error) {
	if u.X == 0 {
		// For zero values, the RLP encoding is a zero-length string.
		return []byte{stringOffset}, nil
	}
	d := make([]byte, 8)
	l := writeInt(d, u.X)
	return NewBytes(d[:l]).EncodeRLP()
}

func (u *UintItem) DecodeRLP(d []byte) (int, error) {
	s := &StringItem{}
	i, err := s.DecodeRLP(d)
	if err != nil {
		return 0, err
	}
	n, err := readInt(*s, len(*s))
	if err != nil {
		return 0, err
	}
	(*u).X = n
	return i, nil
}

// BigIntItem is a RLP encoded big integer.
type BigIntItem struct{ X *big.Int }

func (b BigIntItem) EncodeRLP() ([]byte, error) {
	if b.X == nil || b.X.Sign() == 0 {
		// For zero values, the RLP encoding is a zero-length string.
		return []byte{stringOffset}, nil
	}
	return NewBytes(b.X.Bytes()).EncodeRLP()
}

func (b *BigIntItem) DecodeRLP(d []byte) (int, error) {
	s := &StringItem{}
	i, err := s.DecodeRLP(d)
	if err != nil {
		return 0, err
	}
	(*b).X = new(big.Int).SetBytes(*s)
	return i, nil
}
