package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHeaderRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected Header
	}{
		{
			name: "confirmable GET",
			data: append([]byte{
				0x48,       // Version 1, Confirmable, Token Length 8}
				0x01,       // Code 1 (GET)
				0x42, 0x42, // Message ID 0x4242
			}, bytes8...),
			expected: Header{
				Version:   ProtocolVersion,
				Type:      Confirmable,
				Code:      uint8(Get),
				MessageID: 0x4242,
				Token:     bytes8,
			},
		},
	}
	for _, test := range tests {
		header := Header{}

		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			err := header.UnmarshalBinary(test.data)
			if err != nil {
				t.Fatal("unmarshal:", err)
			}

			diff := cmp.Diff(test.expected, header)
			if diff != "" {
				t.Errorf("header mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(test.name+"/append", func(t *testing.T) {
			data, err := header.AppendBinary(nil)
			if err != nil {
				t.Fatal("append:", err)
			}

			if diff := cmp.Diff(test.data, data); diff != "" {
				t.Errorf("data mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
