package coap

import "fmt"

type ParseError struct {
	Offset int
	Cause  error
}

type TruncatedError struct {
	Expected int
}

type UnsupportedVersion struct {
	Version uint8
}
type InvalidTokenLength struct {
	Length int
}

type UnsupportedExtendError struct{}

type OptionValueLengthError struct {
	OptionDef
	Length uint16
}

type OptionValueFormatError struct {
	OptionDef
	Requested ValueFormat
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

func (e InvalidTokenLength) Error() string {
	return fmt.Sprintf("unsupported token length %d, max is %d", e.Length, TokenMaxLength)
}

func (e UnsupportedExtendError) Error() string {
	return "unsupported extend value"
}

func (e TruncatedError) Error() string {
	return fmt.Sprintf("truncated input, expected %d bytes", e.Expected)
}

func (e OptionValueLengthError) Error() string {
	return fmt.Sprintf("expected option %q value length between %d and %d, got %d", e.Name, e.MinLen, e.MaxLen, e.Length)
}

func (e OptionValueFormatError) Error() string {
	return fmt.Sprintf("unsupported option %q value format %q, actual %q", e.Name, e.Requested, e.ValueFormat)
}
