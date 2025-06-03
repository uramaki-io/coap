package coap

import (
	"encoding/binary"
)

const (
	ExtendByte    = uint8(0x0D) // 13
	ExtendDword   = uint8(0x0E) // 14
	ExtendInvalid = uint8(0x0F) // 15

	ExtendByteOffset  = uint16(ExtendByte)               // 13
	ExtendDwordOffset = uint16(256) + uint16(ExtendByte) // 269
)

type OptionDef struct {
	Name        string
	Code        uint16
	ValueFormat ValueFormat
	MinLen      uint16
	MaxLen      uint16
}

type Option struct {
	OptionDef

	uintValue   uint32
	bytesValue  []byte
	stringValue string
}

type ValueFormat uint8

const (
	ValueFormatEmpty ValueFormat = iota
	ValueFormatUint
	ValueFormatOpaque
	ValueFormatString
)

func (f ValueFormat) String() string {
	switch f {
	case ValueFormatEmpty:
		return "empty"
	case ValueFormatUint:
		return "uint"
	case ValueFormatOpaque:
		return "opaque"
	case ValueFormatString:
		return "string"
	default:
		return "unknown"
	}
}

// Critical returns true if option critical bit is set.
func (o OptionDef) Critical() bool {
	return o.Code&0x01 == 0x01
}

func (o OptionDef) Unsafe() bool {
	return o.Code&0x02 == 0x02
}

func (o OptionDef) NoCacheKey() bool {
	return o.Code&0x1E == 0x1c
}

func (o OptionDef) String() string {
	return o.Name
}

func (o Option) Value() any {
	switch o.ValueFormat {
	case ValueFormatUint:
		return o.uintValue
	case ValueFormatOpaque:
		return o.bytesValue
	case ValueFormatString:
		return o.stringValue
	default:
		return nil
	}
}

func (o Option) GetUint() (uint32, error) {
	if o.ValueFormat != ValueFormatUint {
		return 0, OptionValueFormatError{
			OptionDef: o.OptionDef,
			Requested: ValueFormatUint,
		}
	}

	return o.uintValue, nil
}

func (o *Option) SetUint(value uint32) error {
	if o.ValueFormat != ValueFormatUint {
		return OptionValueFormatError{
			OptionDef: o.OptionDef,
			Requested: ValueFormatUint,
		}
	}

	length := len32(value)
	if length < o.MinLen || length > o.MaxLen {
		return OptionValueLengthError{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	}

	o.uintValue = value

	return nil
}

func (o Option) GetBytes() ([]byte, error) {
	if o.ValueFormat != ValueFormatOpaque {
		return nil, OptionValueFormatError{
			OptionDef: o.OptionDef,
			Requested: ValueFormatOpaque,
		}
	}

	return o.bytesValue, nil
}

func (o *Option) SetBytes(value []byte) error {
	if o.ValueFormat != ValueFormatOpaque {
		return OptionValueFormatError{
			OptionDef: o.OptionDef,
			Requested: ValueFormatOpaque,
		}
	}

	length := uint16(len(value))
	if length < o.MinLen || length > o.MaxLen {
		return OptionValueLengthError{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	}

	o.bytesValue = value

	return nil
}

func (o Option) GetString() (string, error) {
	if o.ValueFormat != ValueFormatString {
		return "", OptionValueFormatError{
			OptionDef: o.OptionDef,
			Requested: ValueFormatString,
		}
	}

	return string(o.stringValue), nil
}

func (o *Option) SetString(value string) error {
	if o.ValueFormat != ValueFormatString {
		return OptionValueFormatError{
			OptionDef: o.OptionDef,
			Requested: ValueFormatString,
		}
	}

	length := uint16(len(value))
	if length < o.MinLen || length > o.MaxLen {
		return OptionValueLengthError{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	}

	o.stringValue = value

	return nil
}

// Encode appends the encoded option to the provided data slice.
func (o Option) Encode(data []byte, prev uint16) ([]byte, error) {
	// determine value length
	length := uint16(0)
	switch o.ValueFormat {
	case ValueFormatUint:
		length = len32(o.uintValue)
	case ValueFormatOpaque:
		length = uint16(len(o.bytesValue))
	case ValueFormatString:
		length = uint16(len(o.stringValue))
	}

	// reserve space for delta/length header
	header := len(data)
	data = append(data, 0)

	// encode delta
	delta := uint16(o.Code - prev)
	hd, data := encodeExtend(data, delta)

	// encode length
	hl, data := encodeExtend(data, length)

	// set delta/length header
	data[header] = hd<<4 | hl

	if length == 0 {
		return data, nil
	}

	switch o.ValueFormat {
	case ValueFormatOpaque:
		data = append(data, o.bytesValue...)
	case ValueFormatString:
		data = append(data, o.stringValue...)
	case ValueFormatUint:
		data = encode32(o.uintValue, data)
	}

	return data, nil
}

func (o *Option) Decode(data []byte, prev uint16, schema *Schema) error {
	if schema == nil {
		panic("schema must not be nil")
	}

	if len(data) == 0 {
		return TruncatedError{
			Expected: 1,
		}
	}

	header := data[0]
	offset := 1

	// decode delta
	delta, offset, err := decodeExtend(data, header>>4, offset)
	if err != nil {
		return err
	}

	// decode length
	length, offset, err := decodeExtend(data, header&0x0F, offset)
	if err != nil {
		return err
	}

	// lookup option definition
	code := prev + delta
	o.OptionDef = schema.OptionDef(code)

	// check length against option definition
	switch {
	case len(data) < offset+int(length):
		return TruncatedError{
			Expected: offset + int(length),
		}
	case length < o.MinLen || length > o.MaxLen:
		return OptionValueLengthError{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	case length == 0:
		return nil
	}

	// decode value
	data = data[offset : offset+int(length)]
	switch o.ValueFormat {
	case ValueFormatOpaque:
		o.bytesValue = data
	case ValueFormatString:
		o.stringValue = string(data)
	case ValueFormatUint:
		o.uintValue = decode32(data)
	}

	return nil
}

func len32(v uint32) uint16 {
	if v == 0 {
		return 0
	}

	switch {
	case v <= 0xFF:
		return 1
	case v <= 0xFFFF:
		return 2
	case v <= 0xFFFFFF:
		return 3
	default:
		return 4
	}
}

func encodeExtend(data []byte, v uint16) (uint8, []byte) {
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

func decodeExtend(data []byte, v uint8, offset int) (uint16, int, error) {
	switch v {
	case ExtendByte:
		if len(data) < offset+1 {
			return 0, offset, TruncatedError{Expected: offset + 1}
		}
		return uint16(data[offset]) + ExtendByteOffset, offset + 1, nil
	case ExtendDword:
		if len(data) < offset+2 {
			return 0, offset, TruncatedError{Expected: offset + 2}
		}
		return binary.BigEndian.Uint16(data[offset:offset+2]) + ExtendDwordOffset, offset + 2, nil
	case ExtendInvalid:
		return 0, offset, UnsupportedExtendError{}
	default:
		return uint16(v), offset, nil
	}
}

// encode32 encodes a uint32 value in big-endian format using the minimum number of bytes
func encode32(v uint32, data []byte) []byte {
	switch {
	case v <= 0xFF:
		return append(data, uint8(v))
	case v <= 0xFFFF:
		return append(data, uint8(v>>8), uint8(v))
	case v <= 0xFFFFFF:
		return append(data, uint8(v>>16), uint8(v>>8), uint8(v))
	default:
		return append(data, uint8(v>>24), uint8(v>>16), uint8(v>>8), uint8(v))
	}
}

// decode32 decodes a uint32 value from big-endian format using the minimum number of bytes
func decode32(data []byte) uint32 {
	switch len(data) {
	case 1:
		return uint32(data[0])
	case 2:
		return uint32(data[0])<<8 | uint32(data[1])
	case 3:
		return uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	case 4:
		return uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	default:
		panic("invalid data length for decode32")
	}
}
