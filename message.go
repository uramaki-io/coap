package coap

import "slices"

// PayloadMarker is the marker byte that indicates the presence of a payload in a CoAP message.
const PayloadMarker = 0xFF

// Message represents a CoAP message, which includes a header, options, and an optional payload.
type Message struct {
	Header
	Options

	Payload []byte
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
	_, err := m.Decode(data, DefaultSchema)
	return err
}

// Decode decodes the CoAP message from the provided data slice using the given schema.
//
// Returns the remaining data after the message and UnarshalError if any error occurs during decoding.
func (m *Message) Decode(data []byte, schema *Schema) ([]byte, error) {
	if schema == nil {
		schema = DefaultSchema
	}

	length := len(data)

	var err error
	data, err = m.Header.Decode(data)
	if err != nil {
		return data, UnmarshalError{
			Offset: uint(length - len(data)),
			Cause:  err,
		}
	}

	data, err = m.Options.Decode(data, schema)
	if err != nil {
		return data, UnmarshalError{
			Offset: uint(length - len(data)),
			Cause:  err,
		}
	}

	// payload exists if marker was present when decoding options
	if len(data) > 1 {
		m.Payload = slices.Clone(data[1:])
		data = data[len(data):]
	}

	return data, nil
}
