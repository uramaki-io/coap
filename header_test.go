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
			data: []byte{
				0x44,       // Version 1, Confirmable, Token Length 4}
				0x01,       // Code 1 (GET)
				0x42, 0x42, // Message ID 0x4242
				0xde, 0xad, 0xbe, 0xef,
			},
			expected: Header{
				Version:   ProtocolVersion,
				Type:      Confirmable,
				Code:      Code(GET),
				MessageID: 0x4242,
				Token:     bytes4,
			},
		},
	}
	for _, test := range tests {
		header := Header{}

		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			data, err := header.Decode(test.data)
			if err != nil {
				t.Fatal("unmarshal:", err)
			}

			if len(data) != 0 {
				t.Errorf("unexpected trailing data: %x", data)
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

func TestCode(t *testing.T) {
	code := Code(MethodNotAllowed)

	if expected := uint8(4); code.Class() != expected {
		t.Errorf("expected class %d, got %d", expected, code.Class())
	}

	if expected := uint8(5); code.Detail() != expected {
		t.Errorf("expected detail %d, got %d", expected, code.Detail())
	}

	if expected := "4.05"; code.String() != expected {
		t.Errorf("expected string %q, got %q", expected, code.String())
	}
}
