package coap

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"
	"strconv"
)

const (
	// ExtendByte indicates that next header byte is an extended delta or length value.
	ExtendByte = uint8(0x0D) // 13

	// ExtendDword indicates that next two header bytes are an extended delta or length value.
	ExtendDword = uint8(0x0E) // 14

	// ExtendInvalid indicates that the extended value is invalid.
	ExtendInvalid = uint8(0x0F) // 15

	// ExtendByteOffset is the offset for extended byte values.
	ExtendByteOffset = uint16(ExtendByte) // 13

	// ExtendDwordOffset is the offset for extended dword values.
	ExtendDwordOffset = uint16(256) + uint16(ExtendByte) // 269
)

// Option represents a CoAP option, which includes its definition and uint/opaque/string value.
type Option struct {
	OptionDef

	uintValue   uint32
	opaqueValue []byte
	stringValue string
}

// Must panics if the provided error is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustValue returns the value if err is nil, otherwise it panics with the error.
func MustValue[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

// MustMakeOption creates an Option from the provided definition and value, panicking if an error occurs.
func MustMakeOption(def OptionDef, value any) Option {
	opt, err := MakeOption(def, value)
	if err != nil {
		panic(err)
	}

	return opt
}

// MakeOption creates an Option from the provided definition and value.
//
// Returns an error if the value does not match the expected format or length defined in OptionDef.
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

// String returns a string representation of the Option, including its name and value.
//
// If the name is empty it uses the code as a string representation.
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

// GetValue returns the value of the option based on its ValueFormat.
//
// Prefer using specific getter methods like GetUint, GetOpaque, or GetString to ensure type safety and avoid reflect overhead.
// Returns nil if the value format is not recognized.
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

// SetValue sets the value of the option based on its ValueFormat.
//
// Prefer using specific setter methods like SetUint, SetOpaque, or SetString to ensure type safety and avoid reflect overhead.
// Returns an error if the value does not match the expected format or length defined in OptionDef.
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

// GetLength returns the encoded length of the option value.
func (o Option) GetLength() uint16 {
	// determine value length
	switch o.ValueFormat {
	case ValueFormatUint:
		return Len32(o.uintValue)
	case ValueFormatOpaque:
		return uint16(len(o.opaqueValue))
	case ValueFormatString:
		return uint16(len(o.stringValue))
	default:
		return 0
	}
}

// GetUint returns the uint32 value of the option.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatUint.
func (o Option) GetUint() (uint32, error) {
	if o.ValueFormat != ValueFormatUint {
		return 0, InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatUint,
		}
	}

	return o.uintValue, nil
}

// SetUint sets the uint32 value of the option.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatUint.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
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

// GetOpaque returns the opaque byte slice value of the option.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatOpaque.
func (o Option) GetOpaque() ([]byte, error) {
	if o.ValueFormat != ValueFormatOpaque {
		return nil, InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatOpaque,
		}
	}

	return o.opaqueValue, nil
}

// SetOpaque sets the opaque byte slice value of the option.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatOpaque.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
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

// GetString returns the string value of the option.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatString.
func (o Option) GetString() (string, error) {
	if o.ValueFormat != ValueFormatString {
		return "", InvalidOptionValueFormat{
			OptionDef: o.OptionDef,
			Requested: ValueFormatString,
		}
	}

	return string(o.stringValue), nil
}

// SetString sets the string value of the option.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatString.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
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
	// reserve space for delta/length header
	header := len(data)
	data = append(data, 0)

	// encode delta
	delta := uint16(o.Code - prev)
	hd, data := EncodeExtend(data, delta)

	// encode length
	length := o.GetLength()
	hl, data := EncodeExtend(data, length)

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

// Decode decodes the option from the provided data slice, using the previous option code and schema.
//
// Returns the remaining data after decoding and any error encountered during decoding.
// Returns TruncatedError if the data is too short to decode the option.
// Returns InvalidOptionValueLength if the decoded length does not match the expected length defined in OptionDef.
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
	delta, data, err = DecodeExtend(data, header>>4)
	if err != nil {
		return data, err
	}

	// decode length
	var length uint16
	length, data, err = DecodeExtend(data, header&0x0F)
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
// Returns the decoded value, the remaining data slice, and an error if any.
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
