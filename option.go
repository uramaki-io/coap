package coap

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"
	"strconv"
)

const (
	ExtendByte    = uint8(0x0D) // 13
	ExtendDword   = uint8(0x0E) // 14
	ExtendInvalid = uint8(0x0F) // 15

	ExtendByteOffset  = uint16(ExtendByte)               // 13
	ExtendDwordOffset = uint16(256) + uint16(ExtendByte) // 269
)

type Option struct {
	OptionDef

	uintValue   uint32
	opaqueValue []byte
	stringValue string
}

type OptionDef struct {
	Name        string
	Code        uint16
	ValueFormat ValueFormat
	Repeatable  bool
	MinLen      uint16
	MaxLen      uint16
}

type ValueFormat uint8

const (
	ValueFormatEmpty  ValueFormat = 0x00
	ValueFormatUint   ValueFormat = 0x01
	ValueFormatOpaque ValueFormat = 0x02
	ValueFormatString ValueFormat = 0x03
)

var valueFormatString = map[ValueFormat]string{
	ValueFormatEmpty:  "empty",
	ValueFormatUint:   "uint",
	ValueFormatOpaque: "opaque",
	ValueFormatString: "string",
}

func (f ValueFormat) String() string {
	s, ok := valueFormatString[f]
	if !ok {
		panic("unknown value format: " + strconv.FormatUint(uint64(f), 10))
	}

	return s
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustValue[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

func MustMakeOption(def OptionDef, value any) Option {
	opt, err := MakeOption(def, value)
	if err != nil {
		panic(err)
	}

	return opt
}

func MakeOption(def OptionDef, value any) (Option, error) {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetValue(value)
	if err != nil {
		return Option{}, err
	}

	return opt, nil
}

func UnrecognizedOptionDef(code uint16) OptionDef {
	return OptionDef{
		Code:        code,
		ValueFormat: ValueFormatOpaque,
		MaxLen:      1034,
	}
}

func (o Option) String() string {
	name := o.Name
	if name == "" {
		name = strconv.FormatUint(uint64(o.Code), 10)
	}

	switch o.ValueFormat {
	case ValueFormatUint:
		return fmt.Sprintf("%s(%d)", name, o.uintValue)
	case ValueFormatOpaque:
		return fmt.Sprintf("%s(%x)", name, o.opaqueValue)
	case ValueFormatString:
		return fmt.Sprintf("%s(%q)", name, o.stringValue)
	default:
		return fmt.Sprintf("Option(%s)", name)
	}
}

func (o Option) GetValue() any {
	switch o.ValueFormat {
	case ValueFormatUint:
		return o.uintValue
	case ValueFormatOpaque:
		return o.opaqueValue
	case ValueFormatString:
		return o.stringValue
	default:
		return nil
	}
}

func (o *Option) SetValue(value any) error {
	switch v := value.(type) {
	case uint32:
		return o.SetUint(v)
	case []byte:
		return o.SetOpaque(v)
	case string:
		return o.SetString(v)
	default:
		return InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Unknown:   reflect.TypeOf(value),
		}
	}
}

func (o Option) GetUint() (uint32, error) {
	if o.ValueFormat != ValueFormatUint {
		return 0, InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatUint,
		}
	}

	return o.uintValue, nil
}

func (o *Option) SetUint(value uint32) error {
	if o.ValueFormat != ValueFormatUint {
		return InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatUint,
		}
	}

	length := Len32(value)
	if length < o.MinLen || length > o.MaxLen {
		return InvalidOptionValueLength{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	}

	o.uintValue = value

	return nil
}

func (o Option) GetOpaque() ([]byte, error) {
	if o.ValueFormat != ValueFormatOpaque {
		return nil, InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatOpaque,
		}
	}

	return o.opaqueValue, nil
}

func (o *Option) SetOpaque(value []byte) error {
	if o.ValueFormat != ValueFormatOpaque {
		return InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatOpaque,
		}
	}

	length := uint16(len(value))
	if length < o.MinLen || length > o.MaxLen {
		return InvalidOptionValueLength{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	}

	o.opaqueValue = value

	return nil
}

func (o Option) GetString() (string, error) {
	if o.ValueFormat != ValueFormatString {
		return "", InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatString,
		}
	}

	return string(o.stringValue), nil
}

func (o *Option) SetString(value string) error {
	if o.ValueFormat != ValueFormatString {
		return InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatString,
		}
	}

	length := uint16(len(value))
	if length < o.MinLen || length > o.MaxLen {
		return InvalidOptionValueLength{
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
		length = Len32(o.uintValue)
	case ValueFormatOpaque:
		length = uint16(len(o.opaqueValue))
	case ValueFormatString:
		length = uint16(len(o.stringValue))
	}

	if length < o.MinLen || length > o.MaxLen {
		return data, InvalidOptionValueLength{
			OptionDef: o.OptionDef,
			Length:    length,
		}
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
		data = append(data, o.opaqueValue...)
	case ValueFormatString:
		data = append(data, o.stringValue...)
	case ValueFormatUint:
		data = Encode32(o.uintValue, data)
	}

	return data, nil
}

func (o *Option) Decode(data []byte, prev uint16, schema *Schema) ([]byte, error) {
	if schema == nil {
		panic("schema must not be nil")
	}

	if len(data) == 0 {
		return data, TruncatedError{
			Expected: 1,
		}
	}

	header := data[0]
	data = data[1:]

	// decode delta
	var delta uint16
	var err error
	delta, data, err = decodeExtend(data, header>>4)
	if err != nil {
		return data, err
	}

	// decode length
	var length uint16
	length, data, err = decodeExtend(data, header&0x0F)
	if err != nil {
		return data, err
	}

	// lookup option definition
	code := prev + delta
	o.OptionDef = schema.Option(code)

	// check length against option definition
	switch {
	case len(data) < int(length):
		return data, TruncatedError{
			Expected: uint(length),
		}
	case length < o.MinLen || length > o.MaxLen:
		return data, InvalidOptionValueLength{
			OptionDef: o.OptionDef,
			Length:    length,
		}
	case length == 0:
		return data, nil
	}

	// decode value
	switch o.ValueFormat {
	case ValueFormatOpaque:
		o.opaqueValue = slices.Clone(data[:length])
	case ValueFormatString:
		o.stringValue = string(data[:length])
	case ValueFormatUint:
		o.uintValue = Decode32(data[:length])
	}

	return data[length:], nil
}

func (o OptionDef) Recognized() bool {
	return o.Name != ""
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

// Len32 returns minimum number of bytes required to encode a uint32 value in big-endian format
func Len32(v uint32) uint16 {
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

// Encode32 encodes a uint32 value in big-endian format using the minimum number of bytes
func Encode32(v uint32, data []byte) []byte {
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

// Decode32 decodes a uint32 value from big-endian format using the minimum number of bytes
func Decode32(data []byte) uint32 {
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

func decodeExtend(data []byte, v uint8) (uint16, []byte, error) {
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
