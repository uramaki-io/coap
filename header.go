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
	return fmt.Sprintf("Code(%d.%02d)", c.Class(), c.Detail())
}

type Type uint8

const (
	Confirmable     Type = 0x00
	NonConfirmable  Type = 0x01
	Acknowledgement Type = 0x02
	Reset           Type = 0x03
)

func (t Type) String() string {
	switch t {
	case Confirmable:
		return "CON"
	case NonConfirmable:
		return "NON"
	case Acknowledgement:
		return "ACK"
	case Reset:
		return "RST"
	default:
		return fmt.Sprintf("Type(%d)", t)
	}
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

	tpe := (b & 0x30) >> 4
	tkl := b & 0x0f

	if tkl > TokenMaxLength {
		return data, UnsupportedTokenLength{
			Length: uint(tkl),
		}
	}

	expected := HeaderLength + uint(tkl)
	if len(data) < int(expected) {
		return data, TruncatedError{
			Expected: expected,
		}
	}

	h.Version = version
	h.Type = Type(tpe)
	h.Code = Code(data[1])
	h.MessageID = MessageID(binary.BigEndian.Uint16(data[2:4]))
	h.Token = Token(data[4 : 4+tkl])

	return data[4+tkl:], nil
}
