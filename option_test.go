package coap

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	bytes8   = []byte("deadbeef")
	bytes16  = slices.Repeat(bytes8, 2)  // length extend byte
	bytes272 = slices.Repeat(bytes8, 34) // length extend dword
)

func TestOptionSetBytes(t *testing.T) {
	tests := []struct {
		name  string
		def   OptionDef
		value []byte
		err   error
	}{
		{
			name:  "valid opaque value",
			def:   IfMatch,
			value: bytes8,
		},
		{
			name:  "not an opaque value",
			def:   UriHost,
			value: bytes8,
			err: OptionValueFormatError{
				OptionDef: UriHost,
				Requested: ValueFormatOpaque,
			},
		},
		{
			name:  "opaque value too long",
			def:   IfMatch,
			value: bytes272,
			err: OptionValueLengthError{
				OptionDef: IfMatch,
				Length:    272,
			},
		},
		{
			name: "opaque value too short",
			def:  ETag,
			err: OptionValueLengthError{
				OptionDef: ETag,
				Length:    0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opt := Option{
				OptionDef: test.def,
			}

			err := opt.SetBytes(test.value)
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOptionRoundtrip(t *testing.T) {

	tests := []struct {
		name     string
		input    []byte
		def      OptionDef
		expected any
	}{
		{
			name:  "empty value format",
			input: []byte{0x50},
			def:   IfNoneMatch,
		},
		{
			name:     "opaque value format",
			input:    []byte{0x14, 0xde, 0xad, 0xbe, 0xef},
			def:      IfMatch,
			expected: []byte{0xde, 0xad, 0xbe, 0xef},
		},
		{
			name:     "string value format",
			input:    append([]byte{0x38}, bytes8...),
			def:      UriHost,
			expected: string(bytes8),
		},
		{
			name:     "uint value format/1",
			input:    []byte{0x71, 0x42},
			def:      UriPort,
			expected: uint32(0x42),
		},
		{
			name:     "uint value format/2",
			input:    []byte{0x72, 0x42, 0x42},
			def:      UriPort,
			expected: uint32(0x4242),
		},
		{
			name:     "uint value format/3",
			input:    []byte{0xD3, 0x01, 0x42, 0x42, 0x42},
			def:      MaxAge,
			expected: uint32(0x424242),
		},
		{
			name:     "uint value format/4",
			input:    []byte{0xD4, 0x01, 0x42, 0x42, 0x42, 0x42},
			def:      MaxAge,
			expected: uint32(0x42424242),
		},
		{
			name:     "delta extend byte",
			input:    []byte{0xD0, 0x01},
			def:      MaxAge,
			expected: uint32(0),
		},
		{
			name:  "delta extend dword",
			input: []byte{0xE0, 0x00, 0x01},
			def: OptionDef{
				Code: 270,
			},
			expected: []byte(nil),
		},
		{
			name:     "length extend byte",
			input:    append([]byte{0xDD, 0x16, 0x03}, bytes16...),
			def:      ProxyUri,
			expected: string(bytes16),
		},
		{
			name:     "length extend dword",
			input:    append([]byte{0xDE, 0x16, 0x00, 0x03}, bytes272...),
			def:      ProxyUri,
			expected: string(bytes272),
		},
	}

	schema := DefaultSchema()

	for _, test := range tests {
		opt := Option{}

		t.Run(test.name+"/decode", func(t *testing.T) {
			err := opt.Decode(test.input, 0, schema)
			if err != nil {
				t.Fatal("decode:", err)
			}

			if test.def.Code != opt.OptionDef.Code {
				t.Errorf("option code mismatch, want %v, got %v", test.def, opt.OptionDef)
			}

			diff := cmp.Diff(test.expected, opt.Value())
			if diff != "" {
				t.Error("option value mismatch (-want +got):\n", diff)
			}
		})

		t.Run(test.name+"/encode", func(t *testing.T) {
			data, err := opt.Encode(nil, 0)
			if err != nil {
				t.Fatal("encode:", err)
			}

			diff := cmp.Diff(test.input, data)
			if diff != "" {
				t.Error("encoded data mismatch (-want +got):\n", diff)
			}
		})
	}
}

func TestOptionDecodeError(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected error
	}{
		{
			name:  "empty input",
			input: []byte{},
			expected: TruncatedError{
				Expected: 1,
			},
		},
		{
			name:  "truncated input",
			input: []byte{0x71},
			expected: TruncatedError{
				Expected: 2,
			},
		},
		{
			name:     "invalid delta",
			input:    []byte{0xF0},
			expected: UnsupportedExtendError{},
		},
		{
			name:  "truncated delta extend byte",
			input: []byte{0xD0},
			expected: TruncatedError{
				Expected: 2,
			},
		},
		{
			name:  "truncated delta extend dword",
			input: []byte{0xE0, 0x01},
			expected: TruncatedError{
				Expected: 3,
			},
		},
		{
			name:     "invalid length",
			input:    []byte{0x7F},
			expected: UnsupportedExtendError{},
		},
		{
			name:  "truncated length extend byte",
			input: []byte{0x7D},
			expected: TruncatedError{
				Expected: 2,
			},
		},
		{
			name:  "truncated length extend dword",
			input: []byte{0x7E},
			expected: TruncatedError{
				Expected: 3,
			},
		},
		{
			name:  "value length",
			input: []byte{0x73, 0x01, 0x02, 0x03},
			expected: OptionValueLengthError{
				OptionDef: UriPort,
				Length:    3,
			},
		},
	}

	schema := DefaultSchema()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opt := Option{}
			err := opt.Decode(test.input, 0, schema)
			if !errors.Is(err, test.expected) {
				t.Errorf("error mismatch, want %v, got %v", err, test.expected)
			}
		})
	}
}
