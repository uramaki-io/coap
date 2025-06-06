package coap

import "slices"

const (
	// MaxMessageLength is the default maximum length of entire message.
	MaxMessageLength = 65535

	// MaxPayloadLength is default maximum length of a payload.
	MaxPayloadLength = 65535 - HeaderLength - 1

	// MaxOptions is the default maximum number of options.
	MaxOptions = 256

	// MaxOptionLength is the default maximum length of an individual option value.
	MaxOptionLength = 1024

	// PayloadMarker is the marker byte that indicates the presence of a payload in a CoAP message.
	PayloadMarker = 0xFF
)

// Message represents a CoAP message, which includes a header, options, and an optional payload.
type Message struct {
	Header
	Options

	Payload []byte
}

// DecodeOptions holds options for encoding a CoAP message.
type DecodeOptions struct {
	// Schema
	Schema *Schema

	// MaxMessageLength is the maximum length of entire message.
	MaxMessageLength uint

	// MaxPayloadLength is the maximum length of payload.
	MaxPayloadLength uint

	// MaxOptions is the maximum number of options to encode.
	MaxOptions uint

	// MaxOptionLength is the maximum size of an individual option.
	MaxOptionLength uint16
}

// MarshalBinary implements encoding.BinaryMarshaler
func (m *Message) MarshalBinary() ([]byte, error) {
	data, err := m.AppendBinary(nil)
	return data, err
}

// AppendBinary implements encoding.BinaryAppender
func (m *Message) AppendBinary(data []byte) ([]byte, error) {
	data, err := m.Header.AppendBinary(data)
	if err != nil {
		return data, err
	}

	data = m.Options.Encode(data)

	if len(m.Payload) != 0 {
		data = append(data, PayloadMarker)
		data = append(data, m.Payload...)
	}

	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (m *Message) UnmarshalBinary(data []byte) error {
	_, err := m.Decode(data, DecodeOptions{})
	return err
}

// Decode decodes the CoAP message from the provided data slice using the given schema.
//
// Returns the remaining data after the message and UnarshalError if any error occurs during decoding.
func (m *Message) Decode(data []byte, opts DecodeOptions) ([]byte, error) {
	if opts.MaxMessageLength == 0 {
		opts.MaxMessageLength = MaxMessageLength
	}

	if opts.MaxPayloadLength == 0 {
		opts.MaxPayloadLength = MaxPayloadLength
	}

	length := len(data)
	if length > int(opts.MaxMessageLength) {
		return data, MessageTooLong{
			Length: uint(length),
		}
	}

	data, err := m.Header.Decode(data)
	if err != nil {
		return data, UnmarshalError{
			Offset: uint(length - len(data)),
			Cause:  err,
		}
	}

	data, err = m.Options.Decode(data, opts)
	if err != nil {
		return data, UnmarshalError{
			Offset: uint(length - len(data)),
			Cause:  err,
		}
	}

	if len(data) == 0 {
		return data, nil // no payload
	}

	if len(data) > int(opts.MaxPayloadLength) {
		return data, PayloadTooLong{
			Length: uint(len(data)),
			Limit:  opts.MaxPayloadLength,
		}
	}

	// payload exists if marker was present when decoding options
	if len(data) > 1 {
		m.Payload = slices.Clone(data[1:])
		data = data[len(data):]
	}

	return data, nil
}
