package coap

import (
	"encoding/binary"
)

const (
	ProtocolVersion = 1
	TokenMaxLength  = 8
	HeaderLength    = 4
)

type Header struct {
	Version   uint8
	Type      MessageType
	Code      uint8
	MessageID uint16
	Token     []byte
}

type MessageType uint8

const (
	Confirmable MessageType = iota
	NonConfirmable
	Acknowledgement
	Reset
)

// AppendBinary implements encoding.BinaryAppender
func (h Header) AppendBinary(data []byte) ([]byte, error) {
	if h.Version != ProtocolVersion {
		return data, UnsupportedVersion{
			Version: h.Version,
		}
	}

	tkl := len(h.Token)
	if tkl > TokenMaxLength {
		return data, InvalidTokenLength{
			Length: tkl,
		}
	}

	b := uint8(h.Version<<6) | uint8(h.Type<<4) | uint8(tkl)
	data = append(data, b)
	data = append(data, h.Code)
	data = binary.BigEndian.AppendUint16(data, h.MessageID)
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
	tpe := (b & 0x30) >> 4
	tkl := b & 0x0f

	if version != ProtocolVersion {
		return data, UnsupportedVersion{
			Version: version,
		}
	}

	if tkl > TokenMaxLength {
		return data, InvalidTokenLength{
			Length: int(tkl),
		}
	}

	if len(data) < HeaderLength+int(tkl) {
		return data, TruncatedError{
			Expected: HeaderLength + int(tkl),
		}
	}

	h.Version = version
	h.Type = MessageType(tpe)
	h.Code = data[1]
	h.MessageID = binary.BigEndian.Uint16(data[2:4])
	h.Token = data[4 : 4+tkl]

	return data[4+tkl:], nil
}
