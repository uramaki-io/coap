package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestPacketRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		packet Packet
	}{
		{
			name: "packet with options",
			data: []byte{
				0x44, 0x01, 0x84, 0x9e, 0x51, 0x55, 0x77, 0xe8,
				0xb2, 0x48, 0x69,
				0x04, 0x54, 0x65, 0x73, 0x74, 0x43, 0x61, 0x3d, 0x31,
			},
			packet: Packet{
				Header: Header{
					Version:   ProtocolVersion,
					Type:      Confirmable,
					Code:      uint8(Get),
					MessageID: 0x849e,
					Token:     []byte{0x51, 0x55, 0x77, 0xe8},
				},
				Options: MakeOptions(
					MustStringOption(UriPath, "Hi"),
					MustStringOption(UriPath, "Test"),
					MustStringOption(UriQuery, "a=1"),
				),
			},
		},
		{
			name: "packet with payload",
			data: []byte{
				0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC,
				0xFF, 0x48, 0x65, 0x6C, 0x6C, 0x6F,
			},
			packet: Packet{
				Header: Header{
					Version:   ProtocolVersion,
					Type:      Acknowledgement,
					Code:      uint8(Content),
					MessageID: 0x13FD,
					Token:     []byte{0xD0, 0xE2, 0x4D, 0xAC},
				},
				Payload: []byte("Hello"),
			},
		},
	}
	for _, test := range tests {
		packet := Packet{}

		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			err := packet.UnmarshalBinary(test.data)
			if err != nil {
				t.Fatal("unmarshal:", err)
			}

			diff := cmp.Diff(test.packet, packet, EquateOptions())
			if diff != "" {
				t.Errorf("packet mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(test.name+"/append", func(t *testing.T) {
			data, err := test.packet.AppendBinary(nil)
			if err != nil {
				t.Fatal("append:", err)
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
