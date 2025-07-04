// See RFC 7252, Section 3 for details on message header structure.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-3

package coap

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"slices"
	"sync/atomic"
)

const (
	// ProtocolVersion is the expected version in message header.
	ProtocolVersion = 1

	// TokenLength is default length of a token in bytes
	TokenLength = 4

	// TokenMaxLength is aximum length of a token in bytes
	TokenMaxLength = 8

	// HeaderLength is the expected length of message header.
	HeaderLength = 4
)

// Header represents the CoAP message header.
type Header struct {
	Version uint8
	Type    Type
	Code    Code
	ID      MessageID
	Token   Token
}

// Code represents request method or response code.
//
// 1-byte value where the first 3 bits represent the class and the last 5 bits represent the detail.
type Code uint8

// Type represents the message type in CoAP.
//
// Requests can be Confirmable or NonConfirmable.
// Responses can be of any type.
type Type uint8

const (
	// Confirmable is a message type that requires an acknowledgment.
	Confirmable Type = 0x00

	// NonConfirmable is a message type that does not require an acknowledgment.
	NonConfirmable Type = 0x01

	// Acknowledgement is a message type used to acknowledge a Confirmable message.
	Acknowledgement Type = 0x02

	// Reset is a message type used to indicate that a message could not be processed.
	Reset Type = 0x03
)

// MessageID represents the unique identifier for a CoAP message.
//
// 2-byte value that is used for message deduplication and retransmission.
// Usually it is sequential.
type MessageID uint16

// MessageIDSource is a source function that generates
type MessageIDSource func() MessageID

// Token represents the unique random token for matching requests and responses with maximum length of 8 bytes.
type Token []byte

// TokenSource is a function that generates a random token of a specified length.
type TokenSource func() Token

// MessageIDSequence returns a MessageIDSource that generates sequential message IDs starting from the specified value.
//
// Uses an atomic counter. Values wrap around when they reach the maximum value of 65535 (0xffff).
func MessageIDSequence(start MessageID) MessageIDSource {
	id := atomic.Uint32{}
	id.Store(uint32(start))

	return func() MessageID {
		return MessageID(id.Add(1))
	}
}

// RandTokenSource returns a TokenSource that generates cryptographically random tokens of the length between 1-8 bytes.
//
// If the length is 0, it defaults to 4 bytes.
// If the length is greater than 8, it defaults to 8 bytes.
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

// EncodeExtend encodes a uint16 value as an extended delta or length value in the CoAP header format.
//
// Returns the encoded header byte and the updated data slice.
func EncodeExtend(data []byte, v uint16) (uint8, []byte) {
	switch {
	case v < ExtendByteOffset:
		return uint8(v), data
	case v < ExtendDwordOffset:
		data = append(data, uint8(v-ExtendByteOffset))
		return ExtendByte, data
	default:
		data = binary.BigEndian.AppendUint16(data, v-ExtendDwordOffset)
		return ExtendDword, data
	}
}

// DecodeExtend decodes an extended delta or length value from the CoAP header format.
//
// Returns the decoded value and the remaining data slice, and an error if any.
//
// Returns TruncatedError if the data is too short for the expected extend type,
//
// Returns UnsupportedExtendError if the extend type is 15.
func DecodeExtend(data []byte, v uint8) (uint16, []byte, error) {
	switch v {
	case ExtendByte:
		if len(data) < 1 {
			return 0, data, TruncatedError{
				Expected: 1,
			}
		}
		return uint16(data[0]) + ExtendByteOffset, data[1:], nil
	case ExtendDword:
		if len(data) < 2 {
			return 0, data, TruncatedError{
				Expected: 2,
			}
		}
		return binary.BigEndian.Uint16(data) + ExtendDwordOffset, data[2:], nil
	case ExtendInvalid:
		return 0, data, UnsupportedExtendError{}
	default:
		return uint16(v), data, nil
	}
}

// AppendBinary encodes the CoAP message header to the provided data slice.
//
// Returns the updated data slice with the header appended and any error encountered during encoding.
//
// Returns an UnsupportedVersion error if the header version does not match the expected ProtocolVersion.
//
// Returns an UnsupportedTokenLength error if the token length exceeds the TokenMaxLength.
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
	data = binary.BigEndian.AppendUint16(data, uint16(h.ID))
	data = append(data, h.Token...)

	return data, nil
}

// Decode decodes the CoAP message header from the provided data slice.
//
// Returns the remaining data after the header and any error encountered during decoding.
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
	h.ID = messageID
	h.Token = Token(slices.Clone(data[:tkl]))

	return data[tkl:], nil
}

// Class indicates the class of the request method or response code represented by the first 3 bits of the code value.
func (c Code) Class() uint8 {
	return uint8((c & 0xe0) >> 5)
}

// Detail indicates the detail of the request method or response code represented by the last 5 bits of the code value.
func (c Code) Detail() uint8 {
	return uint8(c & 0x1f)
}

// String returns a string representation of the Code.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-12.1
func (c Code) String() string {
	return fmt.Sprintf("%d.%02d", c.Class(), c.Detail())
}

var typeString = map[Type]string{
	Confirmable:     "CON",
	NonConfirmable:  "NON",
	Acknowledgement: "ACK",
	Reset:           "RST",
}

// String implements fmt.Stringer.
func (t Type) String() string {
	s, ok := typeString[t]
	if !ok {
		return fmt.Sprintf("Type(%d)", t)
	}

	return s
}

// Hash generates FNV-1a hash of the token.
func (t Token) Hash() uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write(t)

	return hash.Sum64()
}
