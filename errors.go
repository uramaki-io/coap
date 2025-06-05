package coap

import "fmt"

type UnsupportedVersion struct {
	Version uint8
}

type UnsupportedType struct {
	Type Type
}

type UnsupportedCode struct {
	Code Code
}

type UnsupportedTokenLength struct {
	Length uint
}

type ParseError struct {
	Offset uint
	Cause  error
}

type TruncatedError struct {
	Expected uint
}

type UnsupportedExtendError struct{}

type OptionNotFound struct {
	OptionDef
}

type InvalidOptionValueFormat struct {
	OptionDef
	Requested ValueFormat
}

type OptionNotRepeateable struct {
	OptionDef
}

type InvalidOptionValueLength struct {
	OptionDef
	Length uint16
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at offset %d: %v", e.Offset, e.Cause)
}

func (e ParseError) Unwrap() error {
	return e.Cause
}

func (e UnsupportedVersion) Error() string {
	return fmt.Sprintf("unsupported version %d, expected %d", e.Version, ProtocolVersion)
}

func (e UnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type %s", e.Type)
}

func (e UnsupportedCode) Error() string {
	return fmt.Sprintf("unsupported code %s", e.Code)
}

func (e UnsupportedTokenLength) Error() string {
	return fmt.Sprintf("unsupported token length %d, max is %d", e.Length, TokenMaxLength)
}

func (e UnsupportedExtendError) Error() string {
	return "unsupported extend value"
}

func (e TruncatedError) Error() string {
	return fmt.Sprintf("truncated input, expected %d bytes", e.Expected)
}

func (e OptionNotFound) Error() string {
	return fmt.Sprintf("option %q not found", e.Name)
}

func (e InvalidOptionValueLength) Error() string {
	return fmt.Sprintf("expected option %q value length between %d and %d, got %d", e.Name, e.MinLen, e.MaxLen, e.Length)
}

func (e InvalidOptionValueFormat) Error() string {
	return fmt.Sprintf("invalid option %q value format %q, actual %q", e.Name, e.Requested, e.ValueFormat)
}

func (e OptionNotRepeateable) Error() string {
	return fmt.Sprintf("option %q is not repeateable", e.Name)
}
