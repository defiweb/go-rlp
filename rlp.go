package rlp

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"reflect"
)

var (
	ErrUnsupportedType         = errors.New("rlp: unsupported type")
	ErrUnexpectedEndOfData     = errors.New("rlp: unexpected end of data")
	ErrUnexpectedTrailingData  = errors.New("rlp: unexpected trailing data")
	ErrUnexpectedNumberOfItems = errors.New("rlp: unexpected number of list items")
	ErrNonCanonicalEncoding    = errors.New("rlp: non-canonical encoding")
	ErrNilValue                = errors.New("rlp: nil value")
	ErrTooLarge                = errors.New("rlp: value too large")
)

// Encoder is the interface implemented by types that can marshal themselves
// into RLP.
type Encoder interface {
	// EncodeRLP returns the RLP encoding of the value.
	EncodeRLP() ([]byte, error)
}

// Decoder is the interface implemented by types that can unmarshal
// themselves from RLP.
type Decoder interface {
	// DecodeRLP decodes the RLP data and stores the result in the value and
	// returns the number of bytes read. The data may be longer than the encoded
	// value, in which case the remaining data is ignored.
	//
	// Implementations must tolerate trailing data because list items are
	// decoded from the remaining payload of the list.
	DecodeRLP([]byte) (int, error)
}

// Encode encodes the given value into an RLP item.
//
// If src is nil, ErrNilValue is returned.
func Encode(src Encoder) ([]byte, error) {
	if isNil(src) {
		return nil, ErrNilValue
	}
	return src.EncodeRLP()
}

// Decode decodes RLP item and stores the result in the value pointed to
// by dst. It returns the number of bytes read and an error, if any.
//
// The data must contain exactly one RLP item, otherwise ErrUnexpectedTrailingData
// is returned. To decode an item followed by other data, such as a stream of
// concatenated items, use the DecodeRLP method of the destination value, or
// DecodeLazy.
//
// The decoded value may share memory with the input data, so the input data
// must not be modified as long as the decoded value is in use.
//
// If dst is nil, ErrNilValue is returned.
func Decode(src []byte, dst Decoder) (int, error) {
	if isNil(dst) {
		return 0, ErrNilValue
	}
	n, err := dst.DecodeRLP(src)
	if err != nil {
		return 0, err
	}
	if n != len(src) {
		return 0, ErrUnexpectedTrailingData
	}
	return n, nil
}

// DecodeLazy performs lazy decoding of RLP encoded data. It returns an RLP
// type that provides methods for further decoding, the number of bytes read
// and an error, if any.
//
// This method may be useful when the exact format of the data is not known.
//
// The decoded value may share memory with the input data, so the input data
// must not be modified as long as the decoded value is in use.
func DecodeLazy(src []byte) (r RLP, n int, err error) {
	n, err = (&r).DecodeRLP(src)
	return
}

const (
	stringOffset   = 0x80
	listOffset     = 0xc0
	singleByteMax  = 0x7f
	shortStringMax = 0xb7
	longStringMax  = 0xbf
	shortListMax   = 0xf7
)

// encodeBytes encodes a byte slice into RLP string item.
func encodeBytes(src []byte) ([]byte, error) {
	if len(src) == 0 {
		// Empty string.
		return []byte{stringOffset}, nil
	}
	if len(src) == 1 && src[0] < stringOffset {
		// Single byte string in the range [0x00, 0x7F].
		return []byte{src[0]}, nil
	}
	prefix, err := encodePrefix(uint64(len(src)), stringOffset)
	if err != nil {
		return nil, err
	}
	return append(prefix, src...), nil
}

// decodeBytes decodes RLP string item into a byte slice.
func decodeBytes(src []byte, dst *[]byte) (int, error) {
	if len(src) == 0 {
		// The data should not be empty. An empty string is encoded as a single
		// byte 0x80.
		return 0, ErrUnexpectedEndOfData
	}
	if src[0] == stringOffset {
		// The data is an empty string.
		*dst = nil
		return 1, nil
	}
	if src[0] < stringOffset {
		// Single byte in the range [0x00, 0x7F].
		*dst = src[:1]
		return 1, nil
	}
	offset, dataLen, prefixLen, err := decodePrefix(src)
	if err != nil {
		return 0, err
	}
	if offset != stringOffset {
		return 0, ErrUnsupportedType
	}
	totalLen := int(dataLen + uint64(prefixLen))
	if len(src) < totalLen {
		return 0, ErrUnexpectedEndOfData
	}
	*dst = src[prefixLen:totalLen]
	return totalLen, nil
}

// encodeList encodes a slice into an RLP list item.
func encodeList(src []any) ([]byte, error) {
	return encodeTypedList(src)
}

// decodeList decodes RLP list item into a slice.
func decodeList(src []byte, dst *[]any) (int, error) {
	return decodeTypedList(src, dst, func() any { return new(RLP) }, false)
}

// encodeTypedList encodes a slice into the RLP list item.
func encodeTypedList[T any](src []T) ([]byte, error) {
	var buf bytes.Buffer
	for _, item := range src {
		if isNil(item) {
			return nil, ErrNilValue
		}
		switch enc := any(item).(type) {
		case Encoder:
			data, err := enc.EncodeRLP()
			if err != nil {
				return nil, err
			}
			buf.Write(data)
		default:
			return nil, ErrUnsupportedType
		}
	}
	prefix, err := encodePrefix(uint64(buf.Len()), listOffset)
	if err != nil {
		return nil, err
	}
	return append(prefix, buf.Bytes()...), nil
}

// decodeTypedList decodes items of RLP list item into a slice. Items already
// present in the slice are decoded into, and, if grow is true, items found
// after them are appended to the slice. Otherwise, the number of items in the
// data must match the length of the slice.
func decodeTypedList[T any](src []byte, dst *[]T, newItem func() T, grow bool) (int, error) {
	offset, dataLen, prefixLen, err := decodePrefix(src)
	if err != nil {
		return 0, err
	}
	if offset != listOffset {
		return 0, ErrUnsupportedType
	}
	totalLen := int(dataLen + uint64(prefixLen))
	if len(src) < totalLen {
		return 0, ErrUnexpectedEndOfData
	}
	expected := len(*dst)
	data := src[prefixLen:totalLen]
	n := 0
	for ; len(data) > 0; n++ {
		if !grow && n >= expected {
			// The data contains more items than expected.
			return 0, ErrUnexpectedNumberOfItems
		}
		// Reuse the item already in the destination slice, if any.
		// Otherwise, create a new one. Nil items are replaced with new ones to
		// avoid dereferencing a nil pointer during decoding.
		var item T
		reuse := n < expected && !isNil((*dst)[n])
		if reuse {
			item = (*dst)[n]
		} else {
			item = newItem()
		}
		dec, ok := any(item).(Decoder)
		if !ok {
			return 0, ErrUnsupportedType
		}
		itemLen, err := dec.DecodeRLP(data)
		if err != nil {
			return 0, err
		}
		if itemLen <= 0 || itemLen > len(data) {
			// The item must not be empty, otherwise the loop would never end,
			// and it must not exceed the list payload.
			return 0, ErrUnexpectedEndOfData
		}
		switch {
		case reuse:
			// The item is already in the destination slice.
		case n < expected:
			// The item in the destination slice is nil.
			(*dst)[n] = item
		default:
			*dst = append(*dst, item)
		}
		data = data[itemLen:]
	}
	if !grow && n < expected {
		// The data contains fewer items than expected.
		return 0, ErrUnexpectedNumberOfItems
	}
	if n == 0 {
		// The data is an empty list.
		*dst = nil
	}
	return totalLen, nil
}

// encodeString encodes a Go string into RLP string item.
func encodeString(src string) ([]byte, error) {
	return encodeBytes([]byte(src))
}

// decodeString decodes RLP string item into a Go string.
func decodeString(src []byte, dst *string) (int, error) {
	var b []byte
	i, err := decodeBytes(src, &b)
	if err != nil {
		return 0, err
	}
	*dst = string(b)
	return i, nil
}

// encodeUint encodes an Go unsigned integer into RLP integer item.
func encodeUint(src uint64) ([]byte, error) {
	if src == 0 {
		// For zero values, the RLP encoding is a zero-length string.
		return []byte{stringOffset}, nil
	}
	d := make([]byte, 8)
	l := writeInt(d, src)
	return encodeBytes(d[:l])
}

// decodeUint decodes RLP integer item into a Go unsigned integer.
func decodeUint(src []byte, dst *uint64) (int, error) {
	var b []byte
	i, err := decodeBytes(src, &b)
	if err != nil {
		return 0, err
	}
	if len(b) > 8 {
		return 0, ErrTooLarge
	}
	if err := verifyCanonicalInt(b); err != nil {
		return 0, err
	}
	n, err := readInt(b, uint8(len(b)))
	if err != nil {
		return 0, err
	}
	*dst = n
	return i, nil
}

// encodeBigInt encodes a Go big integer into an RLP integer item.
func encodeBigInt(src *big.Int) ([]byte, error) {
	if src.Sign() == 0 {
		// For zero values, the RLP encoding is a zero-length string.
		return []byte{stringOffset}, nil
	}
	return encodeBytes(src.Bytes())
}

// decodeBigInt decodes RLP integer item into a Go big integer.
func decodeBigInt(src []byte, dst *big.Int) (int, error) {
	var b []byte
	i, err := decodeBytes(src, &b)
	if err != nil {
		return 0, err
	}
	if err := verifyCanonicalInt(b); err != nil {
		return 0, err
	}
	dst.SetBytes(b)
	return i, nil
}

// encodePrefix encodes the RLP prefix for given offset and length. The offset
// value must be either stringOffset or listOffset.
func encodePrefix(length uint64, offset byte) ([]byte, error) {
	// For length 0-55, the RLP encoding consists of a single byte with value
	// stringOffset or listOffset plus the length of the data.
	if length <= 55 {
		return []byte{offset + byte(length)}, nil
	}
	// For longer data, the RLP encoding consists of a single byte with value
	// stringOffset or listOffset plus 55 and plus number of bytes required to
	// represent the length of the data.
	prefix := make([]byte, 9)
	bytesLen := writeInt(prefix[1:], length)
	if bytesLen >= 8 {
		return nil, ErrTooLarge
	}
	prefix[0] = offset + byte(bytesLen) + 55
	return prefix[:bytesLen+1], nil
}

// decodePrefix decodes RLP prefix and returns offset, data length, and prefix
// length. Any data after the prefix is ignored.
func decodePrefix(prefix []byte) (offset byte, dataLen uint64, prefixLen uint8, err error) {
	if len(prefix) == 0 {
		return 0, 0, 0, ErrUnexpectedEndOfData
	}
	cur := prefix[0]
	switch {
	case cur <= singleByteMax:
		// For a single byte whose value is in the [0x00, 0x7F] range, that
		// byte is its own RLP encoding.
		offset = stringOffset
		dataLen = 1
	case cur <= shortStringMax:
		// If a string is 0-55 bytes long, the RLP encoding consists of a
		// single byte with value 0x80 plus the length of the string followed
		// by the string. The range of the first byte is thus [0x80, 0xB7].
		offset = stringOffset
		dataLen = uint64(cur - stringOffset)
		prefixLen = 1
		if dataLen == 1 {
			// A single byte in the [0x00, 0x7F] range must be encoded as
			// itself, without the prefix.
			if len(prefix) < 2 {
				return 0, 0, 0, ErrUnexpectedEndOfData
			}
			if prefix[1] <= singleByteMax {
				return 0, 0, 0, ErrNonCanonicalEncoding
			}
		}
	case cur <= longStringMax:
		// If a string is more than 55 bytes long, the RLP encoding consists of
		// a single byte with value 0xB7 plus the length of the string in
		// binary form, followed by the length of the string, followed by the
		// string. The range of the first byte is thus [0xB8, 0xBF].
		bytesLen := cur - shortStringMax
		if bytesLen >= 8 {
			return 0, 0, 0, ErrTooLarge
		}
		dataLen, err = readInt(prefix[1:], bytesLen)
		if err != nil {
			return 0, 0, 0, err
		}
		if err := verifyCanonicalLength(prefix[1:], dataLen); err != nil {
			return 0, 0, 0, err
		}
		offset = stringOffset
		prefixLen = 1 + bytesLen
	case cur <= shortListMax:
		// If the total payload of a list (i.e., the combined length of all its
		// items) is 0-55 bytes long, the RLP encoding consists of a single
		// byte with value 0xC0 plus the length of the list followed by the
		// concatenation of the RLP encodings of the items. The range of the
		// first byte is thus [0xC0, 0xF7].
		offset = listOffset
		dataLen = uint64(cur - listOffset)
		prefixLen = 1
	default:
		// If the total payload of a list is more than 55 bytes long, the RLP
		// encoding consists of a single byte with value 0xF7 plus the length
		// of the payload in binary form, followed by the length of the
		// payload, followed by the concatenation of the RLP encodings of the
		// items. The range of the first byte is thus [0xF8, 0xFF].
		bytesLen := cur - shortListMax
		if bytesLen >= 8 {
			return 0, 0, 0, ErrTooLarge
		}
		dataLen, err = readInt(prefix[1:], bytesLen)
		if err != nil {
			return 0, 0, 0, err
		}
		if err := verifyCanonicalLength(prefix[1:], dataLen); err != nil {
			return 0, 0, 0, err
		}
		offset = listOffset
		prefixLen = 1 + bytesLen
	}
	if dataLen+uint64(prefixLen) >= math.MaxInt {
		return 0, 0, 0, ErrTooLarge
	}
	return
}

// verifyCanonicalInt checks whether the given big endian integer is encoded in
// its canonical form, that is, without leading zero bytes.
func verifyCanonicalInt(b []byte) error {
	if len(b) > 0 && b[0] == 0 {
		return ErrNonCanonicalEncoding
	}
	return nil
}

// verifyCanonicalLength checks whether the length stored in the long form of
// the prefix is encoded in its canonical form, that is, without leading zero
// bytes and only for data longer than 55 bytes.
func verifyCanonicalLength(length []byte, dataLen uint64) error {
	if dataLen <= 55 {
		// Data of that length must be encoded using the short form.
		return ErrNonCanonicalEncoding
	}
	if len(length) > 0 && length[0] == 0 {
		return ErrNonCanonicalEncoding
	}
	return nil
}

// isNil returns true if the given value is a nil interface or a nil pointer.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	//nolint:exhaustive
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// writeInt writes an integer to the given buffer in big endian order.
// The number of bytes written is returned.
func writeInt(b []byte, i uint64) int {
	switch {
	case i < 1<<8:
		b[0] = byte(i)
		return 1
	case i < 1<<16:
		b[0] = byte(i >> 8)
		b[1] = byte(i)
		return 2
	case i < 1<<24:
		b[0] = byte(i >> 16)
		b[1] = byte(i >> 8)
		b[2] = byte(i)
		return 3
	case i < 1<<32:
		b[0] = byte(i >> 24)
		b[1] = byte(i >> 16)
		b[2] = byte(i >> 8)
		b[3] = byte(i)
		return 4
	case i < 1<<40:
		b[0] = byte(i >> 32)
		b[1] = byte(i >> 24)
		b[2] = byte(i >> 16)
		b[3] = byte(i >> 8)
		b[4] = byte(i)
		return 5
	case i < 1<<48:
		b[0] = byte(i >> 40)
		b[1] = byte(i >> 32)
		b[2] = byte(i >> 24)
		b[3] = byte(i >> 16)
		b[4] = byte(i >> 8)
		b[5] = byte(i)
		return 6
	case i < 1<<56:
		b[0] = byte(i >> 48)
		b[1] = byte(i >> 40)
		b[2] = byte(i >> 32)
		b[3] = byte(i >> 24)
		b[4] = byte(i >> 16)
		b[5] = byte(i >> 8)
		b[6] = byte(i)
		return 7
	default:
		b[0] = byte(i >> 56)
		b[1] = byte(i >> 48)
		b[2] = byte(i >> 40)
		b[3] = byte(i >> 32)
		b[4] = byte(i >> 24)
		b[5] = byte(i >> 16)
		b[6] = byte(i >> 8)
		b[7] = byte(i)
		return 8
	}
}

// readInt reads an integer of the given length from the beginning of the given
// slice in big endian order. Any remaining data is ignored.
func readInt(data []byte, length uint8) (uint64, error) {
	if len(data) < int(length) {
		return 0, ErrUnexpectedEndOfData
	}
	b := data[0:length]
	switch length {
	case 0:
		return 0, nil
	case 1:
		return uint64(b[0]), nil
	case 2:
		return uint64(uint16(b[1]) |
			uint16(b[0])<<8), nil
	case 3:
		return uint64(uint32(b[2]) |
			uint32(b[1])<<8 |
			uint32(b[0])<<16), nil
	case 4:
		return uint64(uint32(b[3]) |
			uint32(b[2])<<8 |
			uint32(b[1])<<16 |
			uint32(b[0])<<24), nil
	case 5:
		return uint64(b[4]) |
			uint64(b[3])<<8 |
			uint64(b[2])<<16 |
			uint64(b[1])<<24 |
			uint64(b[0])<<32, nil
	case 6:
		return uint64(b[5]) |
			uint64(b[4])<<8 |
			uint64(b[3])<<16 |
			uint64(b[2])<<24 |
			uint64(b[1])<<32 |
			uint64(b[0])<<40, nil
	case 7:
		return uint64(b[6]) |
			uint64(b[5])<<8 |
			uint64(b[4])<<16 |
			uint64(b[3])<<24 |
			uint64(b[2])<<32 |
			uint64(b[1])<<40 |
			uint64(b[0])<<48, nil
	case 8:
		return uint64(b[7]) |
			uint64(b[6])<<8 |
			uint64(b[5])<<16 |
			uint64(b[4])<<24 |
			uint64(b[3])<<32 |
			uint64(b[2])<<40 |
			uint64(b[1])<<48 |
			uint64(b[0])<<56, nil
	}
	return 0, ErrTooLarge
}
