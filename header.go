package coap

import (
	"encoding/binary"
)

const (
	ProtocolVersion = 1
	TokenMaxLength  = 8
	HeaderMinLength = 4
)

type Header struct {
	Version     uint8
	Type        MessageType
	Code        uint8
	TokenLength uint8
	MessageID   uint16
}

type MessageType uint8

const (
	Confirmable MessageType = iota
	NonConfirmable
	Acknowledgement
	Reset
)

// AppendBinary implement encoding.BinaryAppender
func (h Header) AppendBinary(data []byte) ([]byte, error) {
	if h.Version != ProtocolVersion {
		return data, UnsupportedVersion{
			Version: h.Version,
		}
	}

	if h.TokenLength > TokenMaxLength {
		return data, InvalidTokenLength{
			Length: h.TokenLength,
		}
	}

	b := uint8(h.Version<<6) | uint8(h.Type<<4) | h.TokenLength
	data = append(data, b)
	data = append(data, h.Code)
	data = binary.BigEndian.AppendUint16(data, h.MessageID)

	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (h *Header) UnmarshalBinary(data []byte) error {
	if len(data) < HeaderMinLength {
		return TruncatedError{
			Expected: HeaderMinLength,
		}
	}

	b := data[0]
	version := b >> 6
	tpe := (b & 0xcf) >> 4
	tkl := b & 0x0f

	if version != ProtocolVersion {
		return UnsupportedVersion{
			Version: version,
		}
	}

	if tkl > TokenMaxLength {
		return InvalidTokenLength{
			Length: tkl,
		}
	}

	h.Version = version
	h.Type = MessageType(tpe)
	h.TokenLength = tkl
	h.Code = data[1]
	h.MessageID = binary.BigEndian.Uint16(data[2:4])

	return nil
}
