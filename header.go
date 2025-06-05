package coap

import (
	"encoding/binary"
	"fmt"
)

const (
	ProtocolVersion = 1
	TokenMaxLength  = 8
	HeaderLength    = 4
)

type Header struct {
	Version   uint8
	Type      Type
	Code      Code
	MessageID MessageID
	Token     Token
}

type Code uint8

func (c Code) Class() uint8 {
	return uint8((c & 0xe0) >> 5)
}

func (c Code) Detail() uint8 {
	return uint8(c & 0x1f)
}

func (c Code) String() string {
	return fmt.Sprintf("%d.%02d", c.Class(), c.Detail())
}

type Type uint8

const (
	Confirmable     Type = 0x00
	NonConfirmable  Type = 0x01
	Acknowledgement Type = 0x02
	Reset           Type = 0x03
)

var typeString = map[Type]string{
	Confirmable:     "CON",
	NonConfirmable:  "NON",
	Acknowledgement: "ACK",
	Reset:           "RST",
}

func (t Type) String() string {
	s, ok := typeString[t]
	if !ok {
		return fmt.Sprintf("Type(%d)", t)
	}

	return s
}

// AppendBinary implements encoding.BinaryAppender
func (h Header) AppendBinary(data []byte) ([]byte, error) {
	if h.Version != ProtocolVersion {
		return data, UnsupportedVersion{
			Version: h.Version,
		}
	}

	tkl := uint(len(h.Token))
	if tkl > TokenMaxLength {
		return data, UnsupportedTokenLength{
			Length: tkl,
		}
	}

	b := uint8(h.Version<<6) | uint8(h.Type<<4) | uint8(tkl)
	data = append(data, b)
	data = append(data, uint8(h.Code))
	data = binary.BigEndian.AppendUint16(data, uint16(h.MessageID))
	data = append(data, h.Token...)

	return data, nil
}

func (h *Header) Decode(data []byte) ([]byte, error) {
	if len(data) < HeaderLength {
		return data, TruncatedError{
			Expected: HeaderLength,
		}
	}

	b := data[0]
	version := b >> 6

	if version != ProtocolVersion {
		return data, UnsupportedVersion{
			Version: version,
		}
	}

	tpe := Type((b & 0x30) >> 4)
	code := Code(data[1])
	messageID := MessageID(binary.BigEndian.Uint16(data[2:4]))
	tkl := int(b & 0x0f)

	data = data[HeaderLength:]

	if tkl > TokenMaxLength {
		return data, UnsupportedTokenLength{
			Length: uint(tkl),
		}
	}

	if len(data) < tkl {
		return data, TruncatedError{
			Expected: uint(tkl),
		}
	}

	h.Version = version
	h.Type = tpe
	h.Code = code
	h.MessageID = messageID
	h.Token = Token(data[:tkl])

	return data[tkl:], nil
}
