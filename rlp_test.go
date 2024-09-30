package rlp

import (
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
			data: []byte{0xc8, 0x83, 0x64, 0x6f, 0x67, 0x83, 0x63, 0x61, 0x74},
			dest: ptr(List{}),
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
			if !reflect.DeepEqual(tt.dest, tt.want) {
				t.Errorf("Decode() got = %#v, want %#v", tt.dest, tt.want)
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
		_, _ = Decode(s, &testList{})
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
	l := TypedList[RLP]{}
	n, err := Decode(bytes, &l)
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
