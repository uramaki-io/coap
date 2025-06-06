package coap

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHeaderRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		header Header
		data   []byte
	}{
		{
			name: "confirmable GET request",
			header: Header{
				Version:   ProtocolVersion,
				Type:      Confirmable,
				Code:      Code(GET),
				MessageID: 0x4242,
				Token:     bytes4,
			},
			data: []byte{
				0x44,       // Version 1, Confirmable, Token Length 4}
				0x01,       // Code 1 (GET)
				0x42, 0x42, // Message ID 0x4242
				0xde, 0xad, 0xbe, 0xef,
			},
		},
		{
			name: "reset",
			header: Header{
				Version:   ProtocolVersion,
				Type:      Reset,
				Code:      Code(InternalServerError),
				MessageID: 0x4242,
				Token:     Token{},
			},
			data: []byte{0x70, 0xa0, 0x42, 0x42},
		},
		{
			name: "non-confirmable Created response",
			header: Header{
				Version:   ProtocolVersion,
				Type:      NonConfirmable,
				Code:      Code(Created),
				MessageID: 0x4242,
				Token:     Token{},
			},
			data: []byte{0x50, 0x41, 0x42, 0x42},
		},
	}
	for _, test := range tests {
		t.Run(test.name+"/append", func(t *testing.T) {
			data, err := test.header.AppendBinary(nil)
			if err != nil {
				t.Fatal("append:", err)
			}

			if diff := cmp.Diff(test.data, data, EquateBinary()); diff != "" {
				t.Errorf("data mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			header := Header{}
			data, err := header.Decode(test.data)
			if err != nil {
				t.Fatal("unmarshal:", err)
			}

			if len(data) != 0 {
				t.Errorf("unexpected trailing data: %x", data)
			}

			diff := cmp.Diff(test.header, header)
			if diff != "" {
				t.Errorf("header mismatch (-want +got):\n%s", diff)
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

func TestTypeString(t *testing.T) {
	got := Confirmable.String()
	want := "CON"
	if got != want {
		t.Errorf("Confirmable.String() = %q, want %q", got, want)
	}

	got = Type(99).String()
	want = "Type(99)"
	if got != want {
		t.Errorf("Type(99).String() = %q, want %q", got, want)
	}
}

func EquateBinary() cmp.Option {
	return cmp.Transformer("Hex", func(b []byte) string {
		return fmt.Sprintf("%#v", b)
	})
}

func TestRandTokenSource(t *testing.T) {
	tests := []struct {
		name   string
		length uint
		expect int
	}{
		{"default length", 0, 4},
		{"max length", 8, 8},
		{"over max length", 20, 8},
		{"custom length", 6, 6},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			src := RandTokenSource(test.length)
			token := src()
			if len(token) != test.expect {
				t.Errorf("RandTokenSource(%d) returned token of length %d, want %d", test.length, len(token), test.expect)
			}
		})
	}
}

func TestMessageIDSequence(t *testing.T) {
	start := MessageID(100)
	seq := MessageIDSequence(start)

	// Should increment and wrap at 0xffff
	for i := 0; i < 5; i++ {
		got := seq()
		want := MessageID(uint16(start) + uint16(i) + 1)
		if got != want {
			t.Errorf("MessageIDSequence: got %d, want %d", got, want)
		}
	}

	// Test wrap-around
	start = 0xfffe
	seq = MessageIDSequence(start)
	id1 := seq() // 0xffff
	id2 := seq() // 0x0000 (wraps)
	id3 := seq() // 0x0001

	if id1 != 0xffff {
		t.Errorf("Expected 0xffff, got %04x", id1)
	}
	if id2 != 0x0000 {
		t.Errorf("Expected 0x0000, got %04x", id2)
	}
	if id3 != 0x0001 {
		t.Errorf("Expected 0x0001, got %04x", id3)
	}
}
