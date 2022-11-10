package rlp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		data    Item
		want    []byte
		wantErr bool
	}{
		{
			data: NewString(""),
			want: []byte{0x80},
		},
		{
			data: NewString("a"),
			want: []byte{0x61},
		},
		{
			data: NewString("dog"),
			want: []byte{0x83, 0x64, 0x6f, 0x67},
		},
		{
			data: NewString(strings.Repeat("a", 56)),
			want: append([]byte{
				0x80 + 55 + 1, // 0x80 offset + 55 for strings longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the string
			}, []byte(strings.Repeat("a", 56))...),
		},
		{
			data: NewString(strings.Repeat("a", 256)),
			want: append([]byte{
				0x80 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
		},
		{
			data: NewList(),
			want: []byte{0xc0},
		},
		{
			data: NewList(NewString("dog"), NewString("cat")),
			want: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
		},
		{
			data: genList(56, NewString("a")),
			want: append([]byte{
				0xc0 + 55 + 1, // 0xc0 offset + 55 for list longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the list
			}, []byte(strings.Repeat("a", 56))...),
		},
		{
			data: genList(256, NewString("a")),
			want: append([]byte{
				0xc0 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
		},
		{
			data: NewList(NewList(NewString("dog"), NewString("cat")), NewString("horse")),
			want: []byte{0xcf, 0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74, 0x85, 0x68, 0x6f, 0x72, 0x73, 0x65},
		},
		{
			data:    errItem{},
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeInto(t *testing.T) {
	tests := []struct {
		data    []byte
		dest    Item
		want    Item
		wantErr bool
	}{
		{
			data: []byte{0x80},
			dest: &StringItem{},
			want: NewString(""),
		},
		{
			data: []byte{0x61},
			dest: &StringItem{},
			want: NewString("a"),
		},
		{
			data: []byte{0x83, 0x64, 0x6f, 0x67},
			dest: &StringItem{},
			want: NewString("dog"),
		},
		{
			data: append([]byte{
				0x80 + 55 + 1, // 0x80 offset + 55 for strings longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the string
			}, []byte(strings.Repeat("a", 56))...),
			dest: &StringItem{},
			want: NewString(strings.Repeat("a", 56)),
		},
		{
			data: append([]byte{
				0x80 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
			dest: &StringItem{},
			want: NewString(strings.Repeat("a", 256)),
		},
		{
			data: []byte{0xc0},
			dest: &ListItem{},
			want: &ListItem{},
		},
		{
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: &ListItem{},
			want: NewList(&RLP{0x83, 0x64, 0x6f, 0x67}, &RLP{0x83, 0x63, 0x61, 0x74}),
		},
		{
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: &ListItem{&StringItem{}, &StringItem{}},
			want: NewList(NewString("dog"), NewString("cat")),
		},
		{
			data: append([]byte{
				0xc0 + 55 + 1, // 0xc0 offset + 55 for list longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the list
			}, []byte(strings.Repeat("a", 56))...),
			dest: genList(56, &StringItem{}),
			want: genList(56, NewString("a")),
		},
		{
			data: append([]byte{
				0xc0 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
			dest: genList(256, &StringItem{}),
			want: genList(256, NewString("a")),
		},
		{
			data: []byte{0xcf, 0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74, 0x85, 0x68, 0x6f, 0x72, 0x73, 0x65},
			dest: &ListItem{&ListItem{&StringItem{}, &StringItem{}}, &StringItem{}},
			want: NewList(NewList(NewString("dog"), NewString("cat")), NewString("horse")),
		},
		{
			data:    []byte{0x80},
			dest:    errItem{},
			wantErr: true,
		},
		{
			data:    []byte{0xc0 + 1, 'a'},
			dest:    &ListItem{errItem{}},
			wantErr: true,
		},
		{
			data:    []byte{0x80 + 55 + 2, 0xff, 0xff},
			dest:    &StringItem{},
			wantErr: true,
		},
		{
			data:    []byte{0xc0 + 55 + 2, 0xff, 0xff},
			dest:    &ListItem{},
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			bts, err := DecodeInto(tt.data, tt.dest)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(tt.dest, tt.want) {
				t.Errorf("Decode() got = %v, want %v", tt.dest, tt.want)
			}
			if err == nil && bts != len(tt.data) {
				t.Errorf("Decode() bts = %v, want %v", bts, len(tt.data))
			}
		})
	}
}

func FuzzDecode(f *testing.F) {
	for _, s := range [][]byte{
		{stringOffset},
		{listOffset},
		{singleByteMax},
		{shortStringMax},
		{longStringMax},
		{shortListMax},
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s []byte) {
		_, _ = DecodeInto(s, &typedListItem{})
	})
}

type errItem struct{}

func (e errItem) EncodeRLP() ([]byte, error) {
	return nil, fmt.Errorf("error")
}

func (e errItem) DecodeRLP([]byte) (int, error) {
	return 0, fmt.Errorf("error")
}

type typedListItem []Item

func (t typedListItem) EncodeRLP() ([]byte, error) {
	return NewList(t...).EncodeRLP()
}

func (t *typedListItem) DecodeRLP(bytes []byte) (int, error) {
	l := &ListItem{}
	n, err := DecodeInto(bytes, l)
	if err != nil {
		return n, err
	}
	for _, item := range *l {
		if item.(*RLP).IsString() {
			i, _ := item.(*RLP).GetStringItem()
			*t = append(*t, i)
		} else {
			i, _ := item.(*RLP).GetListItem()
			*t = append(*t, (*typedListItem)(i))
		}
	}
	return n, nil
}

func genList(n int, i Item) Item {
	l := NewList()
	for j := 0; j < n; j++ {
		l.Append(i)
	}
	return l
}
