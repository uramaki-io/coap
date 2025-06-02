package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestOptionDecodeAppend(t *testing.T) {
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
			input:    []byte{0x35, 'h', 'e', 'l', 'l', 'o'},
			def:      UriHost,
			expected: "hello",
		},
		{
			name:     "uint value format, extended delta",
			input:    []byte{0xd1, 0x01, 0x42},
			def:      MaxAge,
			expected: uint32(0x42),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opt := Option{}
			err := opt.Decode(test.input, 0, DefaultSchema)
			if err != nil {
				t.Fatal("decode:", err)
			}

			diff := cmp.Diff(test.def, opt.OptionDef)
			if diff != "" {
				t.Error("option definition mismatch (-want +got):\n", diff)
			}

			diff = cmp.Diff(test.expected, opt.Value())
			if diff != "" {
				t.Error("option value mismatch (-want +got):\n", diff)
			}

			data, err := opt.Append(nil, 0)
			if err != nil {
				t.Fatal("append:", err)
			}

			if diff := cmp.Diff(test.input, data); diff != "" {
				t.Error("option data mismatch (-want +got):\n", diff)
			}
		})
	}
}
