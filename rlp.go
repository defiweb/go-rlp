package rlp

import (
	"errors"
)

var ErrUnsupportedType = errors.New("rlp: unsupported type")
var ErrUnexpectedEndOfData = errors.New("rlp: unexpected end of data")
var ErrTooLarge = errors.New("rlp: value too large")

// Item represents an item that can be encoded and decoded using RLP.
//
// https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/
type Item interface {
	// EncodeRLP returns the RLP encoding of the item.
	EncodeRLP() ([]byte, error)
	// DecodeRLP decodes the given RLP data and returns the number of bytes
	// read. The given data may be longer than the encoded item, in which case
	// the remaining data is ignored.
	DecodeRLP([]byte) (int, error)
}

// Encode encodes the given item into RLP.
func Encode(data Item) ([]byte, error) {
	return data.EncodeRLP()
}

// Decode performs lazy decoding of RLP encoded data. It returns an RLP type
// that provides methods for further decoding, the number of bytes read and an
// error, if any.
//
// Because decoding is performed lazily, the Encode function may not return an
// error if the data is invalid.
func Decode(data []byte) (*RLP, int, error) {
	r := &RLP{}
	n, err := r.DecodeRLP(data)
	if err != nil {
		return nil, 0, err
	}
	return r, n, nil
}

// DecodeInto decode RLP data and stores the result in the value pointed to
// by dest. It returns the number of bytes read and an error, if any.
func DecodeInto(data []byte, dest Item) (int, error) {
	n, err := dest.DecodeRLP(data)
	if err != nil {
		return 0, err
	}
	return n, nil
}

const (
	stringOffset   = 0x80
	listOffset     = 0xc0
	singleByteMax  = 0x7f
	shortStringMax = 0xb7
	longStringMax  = 0xbf
	shortListMax   = 0xf7
)

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
	prefix := make([]byte, 8)
	bytesLen := writeInt(prefix[1:], length)
	if bytesLen >= 8 {
		return nil, ErrTooLarge
	}
	prefix[0] = offset + byte(bytesLen) + 55
	return prefix[:bytesLen+1], nil
}

// decodePrefix decodes RLP prefix and returns offset, data length and prefix
// length. Any data after the prefix is ignored.
func decodePrefix(prefix []byte) (offset byte, dataLen uint64, prefixLen int, err error) {
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
	case cur <= longStringMax:
		// If a string is more than 55 bytes long, the RLP encoding consists of
		// a single byte with value 0xB7 plus the length of the length of the
		// string in binary form, followed by the length of the string, followed
		// by the string. The range of the first byte is thus [0xB8, 0xBF].
		bytesLen := int(cur - shortStringMax)
		dataLen, err = readInt(prefix[1:], bytesLen)
		if err != nil {
			return 0, 0, 0, err
		}
		if bytesLen >= 8 {
			return 0, 0, 0, ErrTooLarge
		}
		offset = stringOffset
		prefixLen = 1 + bytesLen
	case cur <= shortListMax:
		// If the total payload of a list (i.e. the combined length of all its
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
		// of the length of the payload in binary form, followed by the length of
		// the payload, followed by the concatenation of the RLP encodings of
		// the items. The range of the first byte is thus [0xF8, 0xFF].
		bytesLen := int(cur - shortListMax)
		dataLen, err = readInt(prefix[1:], bytesLen)
		if err != nil {
			return 0, 0, 0, err
		}
		if bytesLen >= 8 {
			return 0, 0, 0, ErrTooLarge
		}
		offset = listOffset
		prefixLen = 1 + bytesLen
	}
	return
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

// readInt reads an integer from the given slice in big endian order.
// The leftmost bytes are ignored if the slice is longer than the integer.
func readInt(data []byte, length int) (uint64, error) {
	if len(data) < length {
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
