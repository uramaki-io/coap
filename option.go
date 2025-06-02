package coap

import (
	"encoding/binary"
	"fmt"
)

type Schema struct {
	options map[uint16]OptionDef
}

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

type OptionValueLengthError struct {
	OptionDef
	Length uint16
}

func (e OptionValueLengthError) Error() string {
	return fmt.Sprintf("expected option %q value length between %d and %d, got %d", e.Name, e.MinLen, e.MaxLen, e.Length)
}

type OptionValueFormatError struct {
	OptionDef
	Requested ValueFormat
}

func (e OptionValueFormatError) Error() string {
	return fmt.Sprintf("unsupported option %q value format %q, actual %q", e.Name, e.Requested, e.ValueFormat)
}

var (
	IfMatch       = OptionDef{Code: 1, Name: "IfMatch", ValueFormat: ValueFormatOpaque, MaxLen: 8}
	UriHost       = OptionDef{Code: 3, Name: "UriHost", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 255}
	ETag          = OptionDef{Code: 4, Name: "ETag", ValueFormat: ValueFormatOpaque, MinLen: 1, MaxLen: 8}
	IfNoneMatch   = OptionDef{Code: 5, Name: "IfNoneMatch", ValueFormat: ValueFormatEmpty}
	Observe       = OptionDef{Code: 6, Name: "Observe", ValueFormat: ValueFormatUint, MaxLen: 3}
	UriPort       = OptionDef{Code: 7, Name: "UriPort", ValueFormat: ValueFormatUint, MaxLen: 2}
	LocationPath  = OptionDef{Code: 8, Name: "LocationPath", ValueFormat: ValueFormatString, MaxLen: 255}
	UriPath       = OptionDef{Code: 11, Name: "UriPath", ValueFormat: ValueFormatString, MaxLen: 255}
	ContentFormat = OptionDef{Code: 12, Name: "ContentFormat", ValueFormat: ValueFormatUint, MaxLen: 2}
	MaxAge        = OptionDef{Code: 14, Name: "MaxAge", ValueFormat: ValueFormatUint, MaxLen: 4}
	UriQuery      = OptionDef{Code: 15, Name: "UriQuery", ValueFormat: ValueFormatString, MaxLen: 255}
	Accept        = OptionDef{Code: 17, Name: "Accept", ValueFormat: ValueFormatUint, MaxLen: 2}
	LocationQuery = OptionDef{Code: 20, Name: "LocationQuery", ValueFormat: ValueFormatString, MaxLen: 255}
	Block1        = OptionDef{Code: 27, Name: "Block1", ValueFormat: ValueFormatUint, MaxLen: 3}
	Block2        = OptionDef{Code: 23, Name: "Block2", ValueFormat: ValueFormatUint, MaxLen: 3}
	ProxyUri      = OptionDef{Code: 35, Name: "ProxyUri", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 1034}
	ProxyScheme   = OptionDef{Code: 39, Name: "ProxyScheme", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 255}
	Size1         = OptionDef{Code: 60, Name: "Size1", ValueFormat: ValueFormatUint, MaxLen: 4}
	Size2         = OptionDef{Code: 28, Name: "Size2", ValueFormat: ValueFormatUint, MaxLen: 4}
	NoResponse    = OptionDef{Code: 258, Name: "NoResponse", ValueFormat: ValueFormatUint, MaxLen: 1}
)

var DefaultSchema = NewSchema(
	IfMatch,
	UriHost,
	ETag,
	IfNoneMatch,
	Observe,
	UriPort,
	LocationPath,
	UriPath,
	ContentFormat,
	MaxAge,
	UriQuery,
	Accept,
	LocationQuery,
	Block1,
	Block2,
	ProxyUri,
	ProxyScheme,
	Size1,
	Size2,
	NoResponse,
)

func NewSchema(defs ...OptionDef) *Schema {
	options := make(map[uint16]OptionDef, len(defs))
	for _, def := range defs {
		options[def.Code] = def
	}

	return &Schema{
		options: options,
	}
}

func (s *Schema) OptionDef(code uint16) OptionDef {
	def, ok := s.options[code]
	if !ok {
		return OptionDef{
			Code:        code,
			Name:        fmt.Sprintf("Option(%d)", code),
			ValueFormat: ValueFormatOpaque,
			MinLen:      0,
			MaxLen:      1034,
		}
	}

	return def
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

	o.stringValue = value

	return nil
}

func (o Option) Append(data []byte, prev uint16) ([]byte, error) {
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

	// check length against option definition
	if length < uint16(o.MinLen) || length > uint16(o.MaxLen) {
		return nil, OptionValueLengthError{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	}

	// reserve space for delta/length header
	i := len(data)
	data = append(data, 0)

	// encode delta
	header := uint8(0)
	delta := uint16(o.Code - prev)
	switch {
	case delta <= 12:
		header = uint8(delta << 4)
	// 1 byte extra delta
	case delta <= 269:
		header = 13 << 4
		data = append(data, uint8(delta-13))
	// 2 byte extra delta
	default:
		header = 14 << 4
		data = binary.BigEndian.AppendUint16(data, delta-269)
	}

	// encode length
	switch {
	case length < 12:
		header = header | uint8(length)
	// 1 byte extra length
	case length <= 269:
		header = header | 13
		data = append(data, uint8(length-13))
	// 2 byte extra length
	default:
		header = header | 14
		data = binary.BigEndian.AppendUint16(data, length-269)
	}

	// set delta/length header
	data[i] = header

	if length == 0 {
		return data, nil
	}

	switch o.ValueFormat {
	case ValueFormatOpaque:
		data = append(data, o.bytesValue...)
	case ValueFormatString:
		data = append(data, o.stringValue...)
	case ValueFormatUint:
		// truncate zero bytes
		b := [4]byte{}
		binary.BigEndian.PutUint32(b[:], o.uintValue)
		data = append(data, b[4-length:]...)
	}

	return data, nil
}

func (o *Option) Decode(data []byte, prev uint16, schema *Schema) error {
	if schema == nil {
		schema = DefaultSchema
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
	case length < o.MinLen || length > o.MaxLen:
		return OptionValueLengthError{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	case len(data) < offset+int(length):
		return TruncatedError{
			Expected: offset + int(length),
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
		// truncate zero bytes
		b := [4]byte{}
		copy(b[4-length:], data)
		o.uintValue = binary.BigEndian.Uint32(b[:])
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

func decodeExtend(data []byte, v uint8, offset int) (uint16, int, error) {
	switch v {
	case 13:
		if len(data) < offset+1 {
			return 0, offset, TruncatedError{Expected: offset + 1}
		}
		return uint16(data[offset]) + 13, offset + 1, nil
	case 14:
		if len(data) < offset+2 {
			return 0, offset, TruncatedError{Expected: offset + 2}
		}
		return binary.BigEndian.Uint16(data[offset:offset+2]) + 269, offset + 2, nil
	case 15:
		return 0, offset, UnsupportedExtendError{}
	default:
		return uint16(v), offset, nil
	}
}
