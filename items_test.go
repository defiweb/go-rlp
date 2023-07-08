package rlp

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
)

func TestRLP(t *testing.T) {
	t.Run("empty-string", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x80}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 0 {
			t.Fatalf("expected length 0, got %v", len(rlp))
		}
	})
	t.Run("single-byte", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x00}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 1 {
			t.Fatalf("expected length 1, got %v", len(rlp))
		}
	})
	t.Run("empty-list", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0xC0}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsList() {
			t.Fatalf("expected list, got string")
		}
		if rlp.Length() != 0 {
			t.Fatalf("expected length 0, got %v", len(rlp))
		}
	})
	t.Run("single-item-list", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0xC1, 0x80}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsList() {
			t.Fatalf("expected list, got string")
		}
		if rlp.Length() != 1 {
			t.Fatalf("expected length 1, got %v", len(rlp))
		}
	})
	t.Run("string", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x83, 'a', 'b', 'c'}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 3 {
			t.Fatalf("expected length 3, got %v", len(rlp))
		}
		if _, err := rlp.GetStringItem(); err != nil {
			t.Fatalf("GetStringItem() failed: %v", err)
		}
		s, err := rlp.GetString()
		if err != nil {
			t.Fatalf("GetString() failed: %v", err)
		}
		if s != "abc" {
			t.Fatalf("GetString() returned %q, expected %q", s, "abc")
		}
	})
	t.Run("bytes", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x83, 'a', 'b', 'c'}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 3 {
			t.Fatalf("expected length 3, got %v", len(rlp))
		}
		b, err := rlp.GetBytes()
		if err != nil {
			t.Fatalf("GetBytes() failed: %v", err)
		}
		if string(b) != "abc" {
			t.Fatalf("GetBytes() returned %q, expected %q", b, "abc")
		}
	})
	t.Run("list", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0xC5, 0x80, 0x80, 0x80, 0x80, 0x80}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsList() {
			t.Fatalf("expected list, got string")
		}
		if rlp.Length() != 5 {
			t.Fatalf("expected length 5, got %v", len(rlp))
		}
		if _, err := rlp.GetListItem(); err != nil {
			t.Fatalf("GetListItem() failed: %v", err)
		}
		l, err := rlp.GetList()
		if err != nil {
			t.Fatalf("GetList() failed: %v", err)
		}
		if len(l) != 5 {
			t.Fatalf("GetList() returned %v items, expected %v", len(l), 5)
		}
	})
	t.Run("uint", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x82, 0x01, 0x00}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 2 {
			t.Fatalf("expected length 2, got %v", len(rlp))
		}
		if _, err := rlp.GetUintItem(); err != nil {
			t.Fatalf("GetUintItem() failed: %v", err)
		}
		u, err := rlp.GetUint()
		if err != nil {
			t.Fatalf("GetUint() failed: %v", err)
		}
		if u != 256 {
			t.Fatalf("GetUint() returned %v, expected %v", u, 256)
		}
	})
	t.Run("bigInt", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x82, 0x01, 0x00}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 2 {
			t.Fatalf("expected length 2, got %v", len(rlp))
		}
		if _, err := rlp.GetBigIntItem(); err != nil {
			t.Fatalf("GetBigIntItem() failed: %v", err)
		}
		bi, err := rlp.GetBigInt()
		if err != nil {
			t.Fatalf("GetBigInt() failed: %v", err)
		}
		if bi.Cmp(big.NewInt(256)) != 0 {
			t.Fatalf("GetBigInt() returned %v, expected %v", bi, big.NewInt(256))
		}
	})
	t.Run("get-success", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{'a'}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		i := &StringItem{}
		ok := false
		err := rlp.Get(i, func(Item) { ok = true })
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		if !ok {
			t.Fatalf("Get() did not call callback")
		}
	})
	t.Run("get-failure", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{'a'}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		i := &ListItem{}
		ok := false
		err := rlp.Get(i, func(Item) { ok = true })
		if err == nil {
			t.Fatalf("Get() did not fail")
		}
		if ok {
			t.Fatalf("Get() called callback")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for _, rlp := range []*RLP{
			{},
			{0x80 + 1},
			{0xc0 + 1},
			{0x80 + 56},
			{0xc0 + 56},
			{0x80 + 56, 1},
			{0xc0 + 56, 1},
		} {
			t.Run(fmt.Sprintf("%x", rlp), func(t *testing.T) {
				if _, err := rlp.GetString(); err == nil {
					t.Fatalf("GetString() should have failed")
				}
				if _, err := rlp.GetBytes(); err == nil {
					t.Fatalf("GetBytes() should have failed")
				}
				if _, err := rlp.GetList(); err == nil {
					t.Fatalf("GetList() should have failed")
				}
				if _, err := rlp.GetUint(); err == nil {
					t.Fatalf("GetUint() should have failed")
				}
				if _, err := rlp.GetBigInt(); err == nil {
					t.Fatalf("GetBigInt() should have failed")
				}
			})
		}
	})
	t.Run("decode-empty", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{}, &rlp); err != ErrUnexpectedEndOfData {
			t.Fatalf("expected ErrUnexpectedEndOfData, got %v", err)
		}
	})
	t.Run("decode-broken-string", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0x81}, &rlp); err != ErrUnexpectedEndOfData {
			t.Fatalf("expected ErrUnexpectedEndOfData, got %v", err)
		}
	})
	t.Run("decode-broken-list", func(t *testing.T) {
		var rlp RLP
		if _, err := DecodeTo([]byte{0xC1}, &rlp); err != ErrUnexpectedEndOfData {
			t.Fatalf("expected ErrUnexpectedEndOfData, got %v", err)
		}
	})
}

func TestStringItemEncode(t *testing.T) {
	tests := []struct {
		data string
		want []byte
	}{
		{"", []byte{0x80}},
		{"a", []byte{'a'}},
		{"ab", []byte{0x82, 'a', 'b'}},
		{string(bytes.Repeat([]byte{'a'}, 55)), append([]byte{0xB7}, bytes.Repeat([]byte{'a'}, 55)...)},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			item := NewString(tt.data)
			got, err := Encode(item)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestStringItemDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    string
		wantErr bool
	}{
		{[]byte{0x80}, "", false},
		{[]byte{0x81, 'a'}, "a", false},
		{[]byte{0x82, 'a', 'b'}, "ab", false},
		{[]byte{0x82, 'a', 'b', 'c'}, "ab", false}, // ignore trailing data
		{append([]byte{0xB7}, bytes.Repeat([]byte{'a'}, 55)...), string(bytes.Repeat([]byte{'a'}, 55)), false},
		{[]byte{0x80 + 56}, "", true},
		{[]byte{0x80 + 56, 255}, "", true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var item StringItem
			_, err := DecodeTo(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if item.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, item.String())
			}
		})
	}
}

func TestListItemEncode(t *testing.T) {
	tests := []struct {
		data *ListItem
		want []byte
	}{
		{&ListItem{}, []byte{0xC0}},
		{&ListItem{NewString("a")}, []byte{0xC0 + 1, 'a'}},
		{&ListItem{NewString("a"), NewString("b")}, []byte{0xC0 + 2, 'a', 'b'}},
		{genList(56, NewString("a")).(*ListItem), append([]byte{0xC0 + 56, 56}, bytes.Repeat([]byte{'a'}, 56)...)},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(tt.data)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestListItemDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    *ListItem
		dest    *ListItem
		wantErr bool
	}{
		{[]byte{0xC0}, &ListItem{}, &ListItem{}, false},
		{[]byte{0xC0 + 1, 'a'}, &ListItem{NewString("a")}, &ListItem{&StringItem{}}, false},
		{[]byte{0xC0 + 2, 'a', 'b'}, &ListItem{NewString("a"), NewString("b")}, &ListItem{&StringItem{}, &StringItem{}}, false},
		{[]byte{0xC0 + 2, 'a', 'b', 'c'}, &ListItem{NewString("a"), NewString("b")}, &ListItem{&StringItem{}, &StringItem{}}, false}, // ignore trailing data
		{append([]byte{0xC0 + 56, 56}, bytes.Repeat([]byte{'a'}, 56)...), genList(56, NewString("a")).(*ListItem), genList(56, &StringItem{}).(*ListItem), false},
		{[]byte{}, nil, &ListItem{}, true},
		{[]byte{0xC0 + 56, 1}, nil, &ListItem{}, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			_, err := DecodeTo(tt.data, tt.dest)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if !reflect.DeepEqual(tt.dest, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, tt.dest)
			}
		})
	}
}

func TestUintItemEncode(t *testing.T) {
	tests := []struct {
		data *UintItem
		want []byte
	}{
		{&UintItem{0}, []byte{0x80}},
		{&UintItem{1}, []byte{1}},
		{&UintItem{127}, []byte{0x7F}},
		{&UintItem{128}, []byte{0x81, 0x80}},
		{&UintItem{1 << 8}, []byte{0x82, 0x01, 0x00}},
		{&UintItem{1 << 16}, []byte{0x83, 0x01, 0x00, 0x00}},
		{&UintItem{1 << 24}, []byte{0x84, 0x01, 0x00, 0x00, 0x00}},
		{&UintItem{1 << 32}, []byte{0x85, 0x01, 0x00, 0x00, 0x00, 0x00}},
		{&UintItem{1 << 40}, []byte{0x86, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{&UintItem{1 << 48}, []byte{0x87, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{&UintItem{1 << 56}, []byte{0x88, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{&UintItem{math.MaxUint64}, []byte{0x88, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(tt.data)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestUintItemDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    *UintItem
		wantErr bool
	}{
		{[]byte{0x80}, &UintItem{0}, false},
		{[]byte{1}, &UintItem{1}, false},
		{[]byte{0x7F}, &UintItem{127}, false},
		{[]byte{0x81, 0x80}, &UintItem{128}, false},
		{[]byte{0x82, 0x01, 0x00}, &UintItem{1 << 8}, false},
		{[]byte{0x83, 0x01, 0x00, 0x00}, &UintItem{1 << 16}, false},
		{[]byte{0x84, 0x01, 0x00, 0x00, 0x00}, &UintItem{1 << 24}, false},
		{[]byte{0x85, 0x01, 0x00, 0x00, 0x00, 0x00}, &UintItem{1 << 32}, false},
		{[]byte{0x86, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}, &UintItem{1 << 40}, false},
		{[]byte{0x87, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, &UintItem{1 << 48}, false},
		{[]byte{0x88, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, &UintItem{1 << 56}, false},
		{[]byte{0x88, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, &UintItem{math.MaxUint64}, false},
		{[]byte{0x89, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, nil, true},
		{[]byte{}, nil, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var item UintItem
			_, err := DecodeTo(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if !reflect.DeepEqual(&item, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, &item)
			}
		})
	}
}

func TestBigIntItemEncode(t *testing.T) {
	tests := []struct {
		data *BigIntItem
		want []byte
	}{
		{&BigIntItem{big.NewInt(0)}, []byte{0x80}},
		{&BigIntItem{big.NewInt(1)}, []byte{1}},
		{&BigIntItem{big.NewInt(127)}, []byte{0x7F}},
		{&BigIntItem{big.NewInt(128)}, []byte{0x81, 0x80}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(tt.data)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestBigIntItemDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    *BigIntItem
		wantErr bool
	}{
		{[]byte{0x80}, &BigIntItem{big.NewInt(0)}, false},
		{[]byte{1}, &BigIntItem{big.NewInt(1)}, false},
		{[]byte{0x7F}, &BigIntItem{big.NewInt(127)}, false},
		{[]byte{0x81, 0x80}, &BigIntItem{big.NewInt(128)}, false},
		{[]byte{}, nil, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var item BigIntItem
			_, err := DecodeTo(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if !reflect.DeepEqual(&item, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, &item)
			}
		})
	}
}
