package coap

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"slices"
	"sync/atomic"
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

type Type uint8

const (
	Confirmable     Type = 0x00
	NonConfirmable  Type = 0x01
	Acknowledgement Type = 0x02
	Reset           Type = 0x03
)

type MessageID uint16

type MessageIDSource func() MessageID

const TokenLength = 4

type Token []byte

type TokenSource func() Token

func RandTokenSource(length uint) TokenSource {
	switch {
	case length == 0:
		length = TokenLength
	case length > TokenMaxLength:
		length = TokenMaxLength
	}

	return func() Token {
		token := make(Token, length)
		_, _ = rand.Read(token) // rand.Read never returns an error

		return token
	}
}

func MessageIDSequence(start MessageID) MessageIDSource {
	id := atomic.Uint32{}
	id.Store(uint32(start))

	return func() MessageID {
		return MessageID(id.Add(1))
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
	h.Token = Token(slices.Clone(data[:tkl]))

	return data[tkl:], nil
}

func (c Code) Class() uint8 {
	return uint8((c & 0xe0) >> 5)
}

func (c Code) Detail() uint8 {
	return uint8(c & 0x1f)
}

func (c Code) String() string {
	return fmt.Sprintf("%d.%02d", c.Class(), c.Detail())
}

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
