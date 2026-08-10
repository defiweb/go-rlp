package rlp

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		data    Encoder
		want    []byte
		wantErr bool
	}{
		{
			data: String(""),
			want: []byte{0x80},
		},
		{
			data: String("a"),
			want: []byte{0x61},
		},
		{
			data: String("dog"),
			want: []byte{0x83, 0x64, 0x6f, 0x67},
		},
		{
			data: String(strings.Repeat("a", 56)),
			want: append([]byte{
				0x80 + 55 + 1, // 0x80 offset + 55 for strings longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the string
			}, []byte(strings.Repeat("a", 56))...),
		},
		{
			data: String(strings.Repeat("a", 256)),
			want: append([]byte{
				0x80 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
		},
		{
			data: List{},
			want: []byte{0xc0},
		},
		{
			data: List{String("dog"), String("cat")},
			want: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
		},
		{
			data: List(makeSlice(56, String("a"))),
			want: append([]byte{
				0xc0 + 55 + 1, // 0xc0 offset + 55 for list longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the list
			}, []byte(strings.Repeat("a", 56))...),
		},
		{
			data: List(makeSlice(256, String("a"))),
			want: append([]byte{
				0xc0 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
		},
		{
			data: List{List{String("dog"), String("cat")}, String("horse")},
			want: []byte{0xcf, 0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74, 0x85, 0x68, 0x6f, 0x72, 0x73, 0x65},
		},
		{
			data:    errItem{},
			wantErr: true,
		},
		{
			data:    (*Uint)(nil),
			wantErr: true,
		},
		{
			data:    List{(*Uint)(nil)},
			wantErr: true,
		},
		{
			data:    List{nil},
			wantErr: true,
		},
		{
			data:    TypedList[Uint]{nil},
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

func TestDecodeTo(t *testing.T) {
	tests := []struct {
		data    []byte
		dest    Decoder
		want    Decoder
		wantErr bool
	}{
		{
			data: []byte{0x80},
			dest: ptr(String("")),
			want: ptr(String("")),
		},
		{
			data: []byte{0x61},
			dest: ptr(String("")),
			want: ptr(String("a")),
		},
		{
			data: []byte{0x83, 0x64, 0x6f, 0x67},
			dest: ptr(String("")),
			want: ptr(String("dog")),
		},
		{
			data: append([]byte{
				0x80 + 55 + 1, // 0x80 offset + 55 for strings longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the string
			}, []byte(strings.Repeat("a", 56))...),
			dest: ptr(String("")),
			want: ptr(String(strings.Repeat("a", 56))),
		},
		{
			data: append([]byte{
				0x80 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
			dest: ptr(String("")),
			want: ptr(String(strings.Repeat("a", 256))),
		},
		{
			data: []byte{0xc0},
			dest: ptr(List{}),
			want: ptr(List(nil)),
		},
		{
			// An empty destination expects an empty list.
			data:    []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest:    ptr(List{}),
			wantErr: true,
		},
		{
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: ptr(List{new(RLP), new(RLP)}),
			want: ptr(List{&RLP{0x83, 0x64, 0x6f, 0x67}, &RLP{0x83, 0x63, 0x61, 0x74}}),
		},
		{
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: ptr(List{new(String), new(String)}),
			want: ptr(List{ptr(String("dog")), ptr(String("cat"))}),
		},
		{
			data: append([]byte{
				0xc0 + 55 + 1, // 0xc0 offset + 55 for list longer than 55 bytes + number of bytes to represent the length
				0x38,          // length of the list
			}, []byte(strings.Repeat("a", 56))...),
			dest: ptr(List(makeSlice(56, new(String)))),
			want: ptr(List(makeSlice(56, ptr(String("a"))))),
		},
		{
			data: append([]byte{
				0xc0 + 55 + 2,
				0x01,
				0x00,
			}, []byte(strings.Repeat("a", 256))...),
			dest: ptr(List(makeSlice(256, new(String)))),
			want: ptr(List(makeSlice(256, ptr(String("a"))))),
		},
		{
			data: []byte{0xcf, 0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74, 0x85, 0x68, 0x6f, 0x72, 0x73, 0x65},
			dest: ptr(List{ptr(List{new(String), new(String)}), new(String)}),
			want: ptr(List{ptr(List{ptr(String("dog")), ptr(String("cat"))}), ptr(String("horse"))}),
		},
		{
			data:    []byte{0x80},
			dest:    ptr(errItem{}),
			wantErr: true,
		},
		{
			data:    []byte{0xc0 + 1, 'a'},
			dest:    ptr(List{errItem{}}),
			wantErr: true,
		},
		{
			data:    []byte{0x80 + 55 + 2, 0xff, 0xff},
			dest:    ptr(String("")),
			wantErr: true,
		},
		{
			data:    []byte{0xc0 + 55 + 2, 0xff, 0xff},
			dest:    ptr(List{}),
			wantErr: true,
		},
		{
			data:    []byte{0x80},
			dest:    (*String)(nil),
			wantErr: true,
		},
		{
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: ptr(TypedList[String]{nil, nil}),
			want: ptr(TypedList[String]{ptr(String("dog")), ptr(String("cat"))}),
		},
		{
			data:    []byte{0xc5, 0x83, 0x64, 0x6f, 0x67},
			dest:    ptr(List{new(String), new(String)}),
			wantErr: true,
		},
		{
			data:    []byte{0xc0},
			dest:    ptr(List{new(String)}),
			wantErr: true,
		},
		{
			// The decoded list is longer than the expected one.
			data:    []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest:    ptr(List{new(String)}),
			wantErr: true,
		},
		{
			// An empty destination expects an empty list.
			data:    []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest:    ptr(TypedList[String]{}),
			wantErr: true,
		},
		{
			// A list of an unknown length can be sized using RLP.Length.
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: ptr(make(TypedList[String], RLP{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74}.Length())),
			want: ptr(TypedList[String]{ptr(String("dog")), ptr(String("cat"))}),
		},
		{
			data:    []byte{0x81, 0x61},
			dest:    ptr(String("")),
			wantErr: true,
		},
		{
			data:    append([]byte{0x80 + 55 + 1, 0x37}, []byte(strings.Repeat("a", 55))...),
			dest:    ptr(String("")),
			wantErr: true,
		},
		{
			data:    append([]byte{0x80 + 55 + 2, 0x00, 0x38}, []byte(strings.Repeat("a", 56))...),
			dest:    ptr(String("")),
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			bts, err := Decode(tt.data, tt.dest)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(tt.dest, tt.want) {
				t.Errorf("Decode() got = %#v, want %#v", tt.dest, tt.want)
			}
			if bts != len(tt.data) {
				t.Errorf("Decode() bts = %v, want %v", bts, len(tt.data))
			}
		})
	}
}

func TestDecodeStream(t *testing.T) {
	// The Decode function accepts a single item only, but items that are
	// followed by other data can be decoded using the DecodeRLP method or the
	// DecodeLazy function.
	data := []byte{0x83, 'f', 'o', 'o', 0x83, 'b', 'a', 'r'}
	if _, err := Decode(data, new(String)); !errors.Is(err, ErrUnexpectedTrailingData) {
		t.Fatalf("expected ErrUnexpectedTrailingData, got %v", err)
	}
	var want = []string{"foo", "bar"}
	for i := 0; len(data) > 0; i++ {
		item, n, err := DecodeLazy(data)
		if err != nil {
			t.Fatalf("DecodeLazy() failed: %v", err)
		}
		s, err := item.String()
		if err != nil {
			t.Fatalf("String() failed: %v", err)
		}
		if s.Get() != want[i] {
			t.Fatalf("expected %q, got %q", want[i], s.Get())
		}
		data = data[n:]
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
		_, _ = Decode(s, &testList{})
	})
}

// FuzzDecodeCanonical verifies that only canonically encoded data is accepted,
// that is, that encoding the decoded data returns the original bytes.
func FuzzDecodeCanonical(f *testing.F) {
	for _, s := range [][]byte{
		{listOffset},
		{listOffset + 1, 0x80},
		{listOffset + 2, 0x81, 0x80},
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s []byte) {
		item, n, err := DecodeLazy(s)
		if err != nil {
			return
		}
		// Decode the item and encode it again. Lists are decoded item by item,
		// strings are decoded as byte slices.
		var enc []byte
		if item.IsList() {
			list, err := item.List()
			if err != nil {
				return
			}
			enc, err = Encode(list)
			if err != nil {
				t.Fatalf("Encode() failed: %v", err)
			}
		} else {
			b, err := item.Bytes()
			if err != nil {
				return
			}
			enc, err = Encode(b)
			if err != nil {
				t.Fatalf("Encode() failed: %v", err)
			}
		}
		if !bytes.Equal(enc, s[:n]) {
			t.Fatalf("non-canonical data accepted: got %x, want %x", s[:n], enc)
		}
	})
}

type errItem struct{}

func (e errItem) EncodeRLP() ([]byte, error) {
	return nil, fmt.Errorf("error")
}

func (e errItem) DecodeRLP([]byte) (int, error) {
	return 0, fmt.Errorf("error")
}

type testList []any

func (t testList) EncodeRLP() ([]byte, error) {
	return Encode(List(t))
}

func (t *testList) DecodeRLP(bytes []byte) (int, error) {
	// The structure of the data is not known in advance, hence it is decoded
	// using the RLP type. Note that the Decode function must not be used here,
	// because a nested item may be followed by other items.
	var r RLP
	n, err := r.DecodeRLP(bytes)
	if err != nil {
		return n, err
	}
	l, err := r.List()
	if err != nil {
		return n, err
	}
	for _, item := range l {
		if item.IsString() {
			i, _ := item.String()
			*t = append(*t, i)
		} else {
			i, _ := item.List()
			*t = append(*t, i)
		}
	}
	return n, nil
}

func makeSlice(n int, i any) []any {
	l := make([]any, n)
	for j := 0; j < n; j++ {
		l[j] = i
	}
	return l
}

func ptr[T any](v T) *T {
	return &v
}
