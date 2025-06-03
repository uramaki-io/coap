package coap

const PayloadMarker = 0xFF

type Packet struct {
	Header
	Options

	Payload []byte
}

// AppendBinary implements encoding.BinaryAppender
func (p Packet) AppendBinary(data []byte) ([]byte, error) {
	data, err := p.Header.AppendBinary(data)
	if err != nil {
		return nil, err
	}

	data, err = p.Options.AppendBinary(data)
	if err != nil {
		return nil, err
	}

	if len(p.Payload) != 0 {
		data = append(data, PayloadMarker)
		data = append(data, p.Payload...)
	}

	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (p *Packet) UnmarshalBinary(data []byte) error {
	return p.Decode(data, DefaultSchema)
}

func (p *Packet) Decode(data []byte, schema *Schema) error {
	if schema == nil {
		panic("schema must not be nil")
	}

	length := len(data)

	var err error
	data, err = p.Header.Decode(data)
	if err != nil {
		return ParseError{
			Offset: length - len(data),
			Cause:  err,
		}
	}

	p.Options = Options{}

	data, err = p.Options.Decode(data, schema)
	if err != nil {
		return ParseError{
			Offset: length - len(data),
			Cause:  err,
		}
	}

	if len(data) != 0 {
		p.Payload = data
	}

	return nil
}
