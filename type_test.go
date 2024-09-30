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
		if _, err := Decode([]byte{0x80}, &rlp); err != nil {
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
		if _, err := Decode([]byte{0x00}, &rlp); err != nil {
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
		if _, err := Decode([]byte{0xC0}, &rlp); err != nil {
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
		if _, err := Decode([]byte{0xC1, 0x80}, &rlp); err != nil {
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
		if _, err := Decode([]byte{0x83, 'a', 'b', 'c'}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 3 {
			t.Fatalf("expected length 3, got %v", len(rlp))
		}
		s, err := rlp.String()
		if err != nil {
			t.Fatalf("String() failed: %v", err)
		}
		if s.Get() != "abc" {
			t.Fatalf("String() returned %q, expected %q", s.Get(), "abc")
		}
	})
	t.Run("bytes", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{0x83, 'a', 'b', 'c'}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 3 {
			t.Fatalf("expected length 3, got %v", len(rlp))
		}
		b, err := rlp.Bytes()
		if err != nil {
			t.Fatalf("Bytes() failed: %v", err)
		}
		if string(b.Get()) != "abc" {
			t.Fatalf("Bytes() returned %q, expected %q", b, "abc")
		}
	})
	t.Run("list", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{0xC5, 0x80, 0x80, 0x80, 0x80, 0x80}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsList() {
			t.Fatalf("expected list, got string")
		}
		if rlp.Length() != 5 {
			t.Fatalf("expected length 5, got %v", len(rlp))
		}
		l, err := rlp.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}
		if len(l.Get()) != 5 {
			t.Fatalf("List() returned %v items, expected %v", len(l.Get()), 5)
		}
	})
	t.Run("uint", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{0x82, 0x01, 0x00}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 2 {
			t.Fatalf("expected length 2, got %v", len(rlp))
		}
		u, err := rlp.Uint()
		if err != nil {
			t.Fatalf("Uint() failed: %v", err)
		}
		if u.Get() != 256 {
			t.Fatalf("Uint() returned %v, expected %v", u, 256)
		}
	})
	t.Run("bigInt", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{0x82, 0x01, 0x00}, &rlp); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !rlp.IsString() {
			t.Fatalf("expected string, got list")
		}
		if rlp.Length() != 2 {
			t.Fatalf("expected length 2, got %v", len(rlp))
		}
		bi, err := rlp.BigInt()
		if err != nil {
			t.Fatalf("BigInt() failed: %v", err)
		}
		if bi.Get().Cmp(big.NewInt(256)) != 0 {
			t.Fatalf("BigInt() returned %v, expected %v", bi, big.NewInt(256))
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
				if _, err := rlp.String(); err == nil {
					t.Fatalf("String() should have failed")
				}
				if _, err := rlp.Bytes(); err == nil {
					t.Fatalf("Bytes() should have failed")
				}
				if _, err := rlp.List(); err == nil {
					t.Fatalf("List() should have failed")
				}
				if _, err := rlp.Uint(); err == nil {
					t.Fatalf("Uint() should have failed")
				}
				if _, err := rlp.BigInt(); err == nil {
					t.Fatalf("BigInt() should have failed")
				}
			})
		}
	})
	t.Run("decode-empty", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{}, &rlp); err != ErrUnexpectedEndOfData {
			t.Fatalf("expected ErrUnexpectedEndOfData, got %v", err)
		}
	})
	t.Run("decode-broken-string", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{0x81}, &rlp); err != ErrUnexpectedEndOfData {
			t.Fatalf("expected ErrUnexpectedEndOfData, got %v", err)
		}
	})
	t.Run("decode-broken-list", func(t *testing.T) {
		var rlp RLP
		if _, err := Decode([]byte{0xC1}, &rlp); err != ErrUnexpectedEndOfData {
			t.Fatalf("expected ErrUnexpectedEndOfData, got %v", err)
		}
	})
}

func TestStringEncode(t *testing.T) {
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
			got, err := Encode(String(tt.data))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestStringDecode(t *testing.T) {
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
			var item String
			_, err := Decode(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if item.Get() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, item.Get())
			}
		})
	}
}

func TestBytesEncode(t *testing.T) {
	tests := []struct {
		data []byte
		want []byte
	}{
		{[]byte(""), []byte{0x80}},
		{[]byte("a"), []byte{'a'}},
		{[]byte("ab"), []byte{0x82, 'a', 'b'}},
		{bytes.Repeat([]byte{'a'}, 55), append([]byte{0xB7}, bytes.Repeat([]byte{'a'}, 55)...)},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(Bytes(tt.data))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestBytesDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    []byte
		wantErr bool
	}{
		{[]byte{0x80}, []byte(""), false},
		{[]byte{0x81, 'a'}, []byte("a"), false},
		{[]byte{0x82, 'a', 'b'}, []byte("ab"), false},
		{[]byte{0x82, 'a', 'b', 'c'}, []byte("ab"), false}, // ignore trailing data
		{append([]byte{0xB7}, bytes.Repeat([]byte{'a'}, 55)...), bytes.Repeat([]byte{'a'}, 55), false},
		{[]byte{0x80 + 56}, []byte(""), true},
		{[]byte{0x80 + 56, 255}, []byte(""), true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var item Bytes
			_, err := Decode(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if !bytes.Equal(item.Get(), tt.want) {
				t.Fatalf("expected %q, got %q", tt.want, item.Get())
			}
		})
	}
}

func TestUintEncode(t *testing.T) {
	tests := []struct {
		data uint64
		want []byte
	}{
		{0, []byte{0x80}},
		{1, []byte{1}},
		{127, []byte{0x7F}},
		{128, []byte{0x81, 0x80}},
		{1 << 8, []byte{0x82, 0x01, 0x00}},
		{1 << 16, []byte{0x83, 0x01, 0x00, 0x00}},
		{1 << 24, []byte{0x84, 0x01, 0x00, 0x00, 0x00}},
		{1 << 32, []byte{0x85, 0x01, 0x00, 0x00, 0x00, 0x00}},
		{1 << 40, []byte{0x86, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{1 << 48, []byte{0x87, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{1 << 56, []byte{0x88, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{math.MaxUint64, []byte{0x88, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(Uint(tt.data))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestUintDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    uint64
		wantErr bool
	}{
		{[]byte{0x80}, 0, false},
		{[]byte{1}, 1, false},
		{[]byte{0x7F}, 127, false},
		{[]byte{0x81, 0x80}, 128, false},
		{[]byte{0x82, 0x01, 0x00}, 1 << 8, false},
		{[]byte{0x83, 0x01, 0x00, 0x00}, 1 << 16, false},
		{[]byte{0x84, 0x01, 0x00, 0x00, 0x00}, 1 << 24, false},
		{[]byte{0x85, 0x01, 0x00, 0x00, 0x00, 0x00}, 1 << 32, false},
		{[]byte{0x86, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}, 1 << 40, false},
		{[]byte{0x87, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 1 << 48, false},
		{[]byte{0x88, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 1 << 56, false},
		{[]byte{0x88, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, math.MaxUint64, false},
		{[]byte{0x89, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0, true},
		{[]byte{}, 0, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var item Uint
			_, err := Decode(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if !reflect.DeepEqual(item.Get(), tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, item.Get())
			}
		})
	}
}

func TestBigIntEncode(t *testing.T) {
	tests := []struct {
		data *big.Int
		want []byte
	}{
		{big.NewInt(0), []byte{0x80}},
		{big.NewInt(1), []byte{1}},
		{big.NewInt(127), []byte{0x7F}},
		{big.NewInt(128), []byte{0x81, 0x80}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode((*BigInt)(tt.data))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestBigIntDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    *big.Int
		wantErr bool
	}{
		{[]byte{0x80}, big.NewInt(0), false},
		{[]byte{1}, big.NewInt(1), false},
		{[]byte{0x7F}, big.NewInt(127), false},
		{[]byte{0x81, 0x80}, big.NewInt(128), false},
		{[]byte{}, nil, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var item BigInt
			_, err := Decode(tt.data, &item)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if !reflect.DeepEqual(item.Get(), tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, item.Get())
			}
		})
	}
}

func TestListEncode(t *testing.T) {
	tests := []struct {
		data []any
		want []byte
	}{
		{[]any{}, []byte{0xC0}},
		{[]any{String("a")}, []byte{0xC0 + 1, 'a'}},
		{[]any{String("a"), Bytes("b")}, []byte{0xC0 + 2, 'a', 'b'}},
		{makeSlice(56, String("a")), append([]byte{0xC0 + 56, 56}, bytes.Repeat([]byte{'a'}, 56)...)},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(List(tt.data))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestListDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    []any
		dest    []any
		wantErr bool
	}{
		{[]byte{0xC0}, nil, []any{}, false},
		{[]byte{0xC0 + 1, 'a'}, []any{ptr(String("a"))}, []any{new(String)}, false},
		{[]byte{0xC0 + 1, 'a'}, []any{&RLP{0x61}}, []any{}, false},
		{[]byte{0xC0 + 2, 'a', 'b'}, []any{ptr(String("a")), ptr(Bytes("b"))}, []any{new(String), new(Bytes)}, false},
		{[]byte{0xC0 + 2, 'a', 'b', 'c'}, []any{ptr(String("a")), ptr(String("b"))}, []any{new(String), new(String)}, false}, // ignore trailing data
		{append([]byte{0xC0 + 56, 56}, bytes.Repeat([]byte{'a'}, 56)...), makeSlice(56, ptr(String("a"))), makeSlice(56, new(String)), false},
		{[]byte{}, nil, []any{}, true},
		{[]byte{0xC0 + 56, 1}, nil, []any{}, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			_, err := Decode(tt.data, (*List)(&tt.dest))
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

func TestTypedListEncode(t *testing.T) {
	tests := []struct {
		data []*String
		want []byte
	}{
		{[]*String{}, []byte{0xC0}},
		{[]*String{ptr(String("a"))}, []byte{0xC0 + 1, 'a'}},
		{[]*String{ptr(String("a")), ptr(String("b"))}, []byte{0xC0 + 2, 'a', 'b'}},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got, err := Encode(TypedList[String](tt.data))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("expected %x, got %x", tt.want, got)
			}
		})
	}
}

func TestTypedListDecode(t *testing.T) {
	tests := []struct {
		data    []byte
		want    []*String
		dest    []*String
		wantErr bool
	}{
		{[]byte{0xC0}, nil, []*String{}, false},
		{[]byte{0xC0 + 1, 'a'}, []*String{ptr(String("a"))}, []*String{new(String)}, false},
		{[]byte{0xC0 + 2, 'a', 'b'}, []*String{ptr(String("a")), ptr(String("b"))}, []*String{new(String), new(String)}, false},
		{[]byte{0xC0 + 2, 'a', 'b', 'c'}, []*String{ptr(String("a")), ptr(String("b"))}, []*String{new(String), new(String)}, false}, // ignore trailing data
		{[]byte{}, nil, []*String{}, true},
		{[]byte{0xC0 + 56, 1}, nil, []*String{}, true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			_, err := Decode(tt.data, (*TypedList[String])(&tt.dest))
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
