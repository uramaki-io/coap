package coap

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	bytes4   = []byte{0xde, 0xad, 0xbe, 0xef} // example opaque value
	bytes8   = slices.Repeat(bytes4, 2)       // example opaque value
	bytes16  = slices.Repeat(bytes8, 2)       // length extend byte
	bytes272 = slices.Repeat(bytes8, 34)      // length extend dword
)

func TestOptionRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		option OptionDef
		data   []byte
		value  any
	}{
		{
			name:   "empty value format",
			option: IfNoneMatch,
			data:   []byte{0x50},
		},
		{
			name:   "opaque value format",
			option: IfMatch,
			data:   append([]byte{0x14}, bytes4...),
			value:  bytes4,
		},
		{
			name:   "string value format",
			option: URIHost,
			data:   append([]byte{0x38}, bytes8...),
			value:  string(bytes8),
		},
		{
			name:   "uint value format/1",
			option: URIPort,
			data:   []byte{0x71, 0x42},
			value:  uint32(0x42),
		},
		{
			name:   "uint value format/2",
			option: URIPort,
			data:   []byte{0x72, 0x42, 0x42},
			value:  uint32(0x4242),
		},
		{
			name:   "uint value format/3",
			option: MaxAge,
			data:   []byte{0xD3, 0x01, 0x42, 0x42, 0x42},
			value:  uint32(0x424242),
		},
		{
			name:   "uint value format/4",
			option: MaxAge,
			data:   []byte{0xD4, 0x01, 0x42, 0x42, 0x42, 0x42},
			value:  uint32(0x42424242),
		},
		{
			name:   "delta extend byte",
			option: MaxAge,
			data:   []byte{0xD0, 0x01},
			value:  uint32(0),
		},
		{
			name: "delta extend dword",
			option: OptionDef{
				Code: 270,
			},
			data:  []byte{0xE0, 0x00, 0x01},
			value: []byte(nil),
		},
		{
			name:   "length extend byte",
			option: ProxyURI,
			data:   append([]byte{0xDD, 0x16, 0x03}, bytes16...),
			value:  string(bytes16),
		},
		{
			name:   "length extend dword",
			option: ProxyURI,
			data:   append([]byte{0xDE, 0x16, 0x00, 0x03}, bytes272...),
			value:  string(bytes272),
		},
		{
			name:   "unrecognized option",
			option: UnrecognizedOptionDef(0xFFFF, MaxOptionLength),
			data:   []byte{0xE0, 0xFE, 0xF2},
			value:  []byte(nil),
		},
	}

	for _, test := range tests {
		opt := Option{}

		t.Run(test.name+"/decode", func(t *testing.T) {
			data, err := opt.Decode(test.data, 0, DecodeOptions{})
			if err != nil {
				t.Fatal("decode:", err)
			}

			if len(data) != 0 {
				t.Errorf("unexpected trailing data: %x", data)
			}

			if test.option.Code != opt.OptionDef.Code {
				t.Errorf("option code mismatch, want %v, got %v", test.option, opt.OptionDef)
			}

			diff := cmp.Diff(test.value, opt.GetValue())
			if diff != "" {
				t.Error("option value mismatch (-want +got):\n", diff)
			}
		})

		t.Run(test.name+"/encode", func(t *testing.T) {
			data := opt.Encode(nil, 0)
			diff := cmp.Diff(test.data, data)
			if diff != "" {
				t.Error("encoded data mismatch (-want +got):\n", diff)
			}
		})
	}
}

func TestOptionSetValue(t *testing.T) {
	tests := []struct {
		name   string
		option OptionDef
		value  any
		err    error
	}{
		{
			name:   "unsupported value",
			option: URIPort,
			value:  42.0,
			err: InvalidOptionValueFormat{
				OptionDef: URIPort,
				Unknown:   reflect.TypeOf(42.0),
			},
		},
		{
			name:   "valid uint value",
			option: URIPort,
			value:  uint32(0x4242),
		},
		{
			name:   "not a uint value",
			option: URIPort,
			value:  "not a uint",
			err: InvalidOptionValueFormat{
				OptionDef: URIPort,
				Requested: ValueFormatString,
			},
		},
		{
			name:   "uint value too long",
			option: URIPort,
			value:  uint32(0x42424242),
			err: InvalidOptionValueLength{
				OptionDef: URIPort,
				Length:    4,
			},
		},
		{
			name:   "valid opaque value",
			option: IfMatch,
			value:  bytes8,
		},
		{
			name:   "not a opaque value",
			option: URIHost,
			value:  bytes8,
			err: InvalidOptionValueFormat{
				OptionDef: URIHost,
				Requested: ValueFormatOpaque,
			},
		},
		{
			name:   "opaque value too long",
			option: IfMatch,
			value:  bytes272,
			err: InvalidOptionValueLength{
				OptionDef: IfMatch,
				Length:    272,
			},
		},
		{
			name:   "opaque value too short",
			option: ETag,
			value:  []byte{},
			err: InvalidOptionValueLength{
				OptionDef: ETag,
				Length:    0,
			},
		},
		{
			name:   "valid string value",
			option: URIHost,
			value:  "example.com",
		},
		{
			name:   "not a string value",
			option: URIHost,
			value:  bytes8,
			err: InvalidOptionValueFormat{
				OptionDef: URIHost,
				Requested: ValueFormatOpaque,
			},
		},
		{
			name:   "string value too short",
			option: URIHost,
			value:  "",
			err: InvalidOptionValueLength{
				OptionDef: URIHost,
				Length:    0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opt := Option{
				OptionDef: test.option,
			}

			err := opt.SetValue(test.value)
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOptionDecodeError(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		err   error
	}{
		{
			name:  "empty input",
			input: []byte{},
			err: TruncatedError{
				Expected: 1,
			},
		},
		{
			name:  "invalid delta",
			input: []byte{0xF0},
			err:   UnsupportedExtendError{},
		},
		{
			name:  "truncated delta extend byte",
			input: []byte{0xD0},
			err: TruncatedError{
				Expected: 1,
			},
		},
		{
			name:  "truncated delta extend dword",
			input: []byte{0xE0, 0x01},
			err: TruncatedError{
				Expected: 2,
			},
		},
		{
			name:  "invalid length",
			input: []byte{0x7F},
			err:   UnsupportedExtendError{},
		},
		{
			name:  "truncated length extend byte",
			input: []byte{0x7D},
			err: TruncatedError{
				Expected: 1,
			},
		},
		{
			name:  "truncated length extend dword",
			input: []byte{0x7E},
			err: TruncatedError{
				Expected: 2,
			},
		},
		{
			name:  "value length",
			input: []byte{0x73, 0x01, 0x02, 0x03},
			err: InvalidOptionValueLength{
				OptionDef: URIPort,
				Length:    3,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opt := Option{}
			_, err := opt.Decode(test.input, 0, DecodeOptions{})
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func expectErr(t testing.TB, err error, expected error) {
	t.Helper()

	diff := cmp.Diff(expected, err, cmpopts.EquateErrors())
	if diff != "" {
		t.Errorf("error mismatch (-want +got):\n%s", diff)
	}
}
