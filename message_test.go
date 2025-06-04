package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestMessageRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		msg  *Message
	}{
		{
			name: "message with options",
			data: []byte{
				0x44, 0x01, 0x84, 0x9e, 0x51, 0x55, 0x77, 0xe8, // Header
				0xb2, 0x48, 0x69, // URIPath "Hi"
				0x04, 0x54, 0x65, 0x73, 0x74, // URIPath "Test"
				0x43, 0x61, 0x3d, 0x31, // URIQuery "a=1"
			},
			msg: &Message{
				Header: Header{
					Version:   ProtocolVersion,
					Type:      Confirmable,
					Code:      Code(GET),
					MessageID: 0x849e,
					Token:     []byte{0x51, 0x55, 0x77, 0xe8},
				},
				Options: MakeOptions(
					MustValue(MakeOption(URIPath, "Hi")),
					MustValue(MakeOption(URIPath, "Test")),
					MustValue(MakeOption(URIQuery, "a=1")),
				),
			},
		},
		{
			name: "message with payload",
			data: []byte{
				0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC, // Header
				0xFF, 0x48, 0x65, 0x6C, 0x6C, 0x6F, // Payload "Hello"
			},
			msg: &Message{
				Header: Header{
					Version:   ProtocolVersion,
					Type:      Acknowledgement,
					Code:      Code(Content),
					MessageID: 0x13FD,
					Token:     []byte{0xD0, 0xE2, 0x4D, 0xAC},
				},
				Payload: []byte("Hello"),
			},
		},
		{
			name: "message with payload and options",
			data: []byte{
				0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC, // Header
				0xD3, 0x01, 0x42, 0x42, 0x42, // MaxAge 0x424242
				0xFF, 0x48, 0x65, 0x6C, 0x6C, 0x6F, // Payload "Hello"
			},
			msg: &Message{
				Header: Header{
					Version:   ProtocolVersion,
					Type:      Acknowledgement,
					Code:      Code(Content),
					MessageID: 0x13FD,
					Token:     []byte{0xD0, 0xE2, 0x4D, 0xAC},
				},
				Options: MakeOptions(
					MustValue(MakeOption(MaxAge, 0x424242)),
				),
				Payload: []byte("Hello"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			msg := &Message{}
			err := msg.UnmarshalBinary(test.data)
			if err != nil {
				t.Fatal("unmarshal:", err)
			}

			diff := cmp.Diff(test.msg, msg, EquateOptions())
			if diff != "" {
				t.Errorf("message mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(test.name+"/marshal", func(t *testing.T) {
			data, err := test.msg.MarshalBinary()
			if err != nil {
				t.Fatal("marshal:", err)
			}

			diff := cmp.Diff(test.data, data)
			if diff != "" {
				t.Errorf("data mismatch: (-want +got):\n%s", diff)
			}
		})
	}
}

func EquateOptions() cmp.Option {
	return cmp.Options{
		cmp.Transformer("Options", func(o Options) []string {
			var opts []string
			for _, opt := range o.data {
				opts = append(opts, opt.String())
			}
			return opts
		}),
		cmpopts.IgnoreUnexported(Options{}),
	}
}
