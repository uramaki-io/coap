package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func FuzzMessageDecode(f *testing.F) {
	f.Add([]byte{0x44, 0x01, 0x84, 0x9e, 0x51, 0x55, 0x77, 0xe8}) // Valid header
	f.Add([]byte{0x70, 0xa0, 0x42, 0x42})                         // Reset header
	f.Add([]byte{0x44, 0x01, 0x00, 0x01, 0xD0, 0xE2, 0x4D})       // Truncated header

	// ensure values are within valid ranges and there is no panic
	f.Fuzz(func(t *testing.T, data []byte) {
		msg := Message{}
		opts := DecodeOptions{
			MaxPayloadLength: 16,
		}
		_, err := msg.Decode(data, opts)
		if err != nil {
			t.SkipNow()
		}

		// 2 bits
		if msg.Version > 1 {
			t.Errorf("invalid version %d", msg.Version)
		}

		// 2 bits
		if msg.Type > 3 {
			t.Errorf("invalid type %d", msg.Type)
		}

		// 5 bits
		if len(msg.Token) > 8 {
			t.Errorf("token length %d exceeds maximum of 8 bytes", len(msg.Token))
		}

		if len(msg.Payload) > int(opts.MaxPayloadLength) {
			t.Errorf("payload length exceeds maximum of %d bytes", opts.MaxPayloadLength)
		}

		for _, opt := range msg.Options {
			if opt.Length() > MaxOptionLength {
				t.Errorf("option value length %d exceeds maximum of %d bytes", opt.Length(), MaxOptionLength)
			}
		}

		if len(msg.Options) > MaxOptions {
			t.Errorf("number of options %d exceeds maximum of %d", len(msg.Options), MaxOptions)
		}
	})
}

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
				Options: Options{
					MustOptionValue(URIPath, "Hi"),
					MustOptionValue(URIPath, "Test"),
					MustOptionValue(URIQuery, "a=1"),
				},
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
				Options: Options{
					MustOptionValue(MaxAge, uint32(0x424242)),
				},
				Payload: []byte("Hello"),
			},
		},
	}
	for _, test := range tests {
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

	}
}

func TestMessageDecodeError(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		opts DecodeOptions
		err  error
	}{
		{
			name: "unknown version",
			data: []byte{0x84, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC},
			err: UnmarshalError{
				Offset: 0,
				Cause: UnsupportedVersion{
					Version: 2,
				},
			},
		},
		{
			name: "unsupported token length",
			data: []byte{0x6c, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC, 0x4D, 0xAC},
			err: UnmarshalError{
				Offset: 4,
				Cause: UnsupportedTokenLength{
					Length: 12,
				},
			},
		},
		{
			name: "truncated header",
			data: []byte{0x64, 0x45},
			err: UnmarshalError{
				Offset: 0,
				Cause: TruncatedError{
					Expected: 4,
				},
			},
		},
		{
			name: "truncated token",
			data: []byte{0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2},
			err: UnmarshalError{
				Offset: 4,
				Cause: TruncatedError{
					Expected: 4,
				},
			},
		},
		{
			name: "truncated options",
			data: []byte{
				0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC, // Header
				0xD3, 0x01, 0x42, // Truncated MaxAge
			},
			err: UnmarshalError{
				Offset: 10,
				Cause: TruncatedError{
					Expected: 3,
				},
			},
		},
		{
			name: "message too long",
			data: []byte{
				0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC, // Header
				0xFF, 0x48, 0x65, 0x6C, 0x6C, 0x6F, // Payload "Hello"
			},
			opts: DecodeOptions{
				MaxMessageLength: 10,
			},
			err: MessageTooLong{
				Limit:  10,
				Length: 14,
			},
		},
		{
			name: "payload too long",
			data: []byte{
				0x64, 0x45, 0x13, 0xFD, 0xD0, 0xE2, 0x4D, 0xAC, // Header
				0xFF, 0x48, 0x65, 0x6C, 0x6C, 0x6F, // Payload "Hello"
			},
			opts: DecodeOptions{
				MaxPayloadLength: 2,
			},
			err: PayloadTooLong{
				Limit:  2,
				Length: 5,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg := &Message{}
			_, err := msg.Decode(test.data, test.opts)

			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMessageMarshalError(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
		err  error
	}{
		{
			name: "unsupported version",
			msg: &Message{
				Header: Header{
					Version: 2,
				},
			},
			err: UnsupportedVersion{
				Version: 2,
			},
		},
		{
			name: "unsupported token length",
			msg: &Message{
				Header: Header{
					Version: ProtocolVersion,
					Token:   bytes16,
				},
			},
			err: UnsupportedTokenLength{
				Length: 16,
			},
		},
	}
	for _, test := range tests {
		_, err := test.msg.MarshalBinary()

		diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
		if diff != "" {
			t.Errorf("error mismatch (-want +got):\n%s", diff)
		}
	}
}
