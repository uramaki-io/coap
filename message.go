package coap

const PayloadMarker = 0xFF

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

	data, err = m.Options.AppendBinary(data)
	if err != nil {
		return data, err
	}

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

func (m *Message) Decode(data []byte, schema *Schema) ([]byte, error) {
	if schema == nil {
		panic("schema must not be nil")
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
		m.Payload = data[1:]
		data = data[len(data):]
	}

	return data, nil
}
