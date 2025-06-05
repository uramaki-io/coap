package coap

import (
	"fmt"
	"reflect"
)

// UnsupportedVersion is returned when the version does not match the expected protocol version 1.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-3
type UnsupportedVersion struct {
	Version uint8
}

// InvalidType is returned when the message type is outside the specified range of 0-4.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-3
type InvalidType struct {
	Type Type
}

// InvalidCode is returned when the code does not match the request/response type.
type InvalidCode struct {
	Code Code
}

// UnsupportedTokenLength is returned when the token length exceeds the maximum allowed length of 8 bytes.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-3
type UnsupportedTokenLength struct {
	Length uint
}

// UnmarshalError is returned when an error occurs during unmarshaling a message.
type UnmarshalError struct {
	// Offset indicates where the error occurred in the input data.
	Offset uint

	// Cause is the underlying error that caused the unmarshaling to fail.
	Cause error
}

// TruncatedError is returned when the input data does not contain enough bytes.
type TruncatedError struct {
	Expected uint
}

// UnsupportedExtendError is returned when an unsupported extend value 15 is encountered.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-3.1
type UnsupportedExtendError struct{}

// OptionNotFound is returned when a requested option is not found in the message options.
type OptionNotFound struct {
	OptionDef
}

// InvalidOptionValueFormat is returned when the value format of an option does not match the requested format.
type InvalidOptionValueFormat struct {
	OptionDef
	Requested ValueFormat
	Unknown   reflect.Type
}

// OptionNotRepeateable is returned when an option that is not allowed to be repeated is found more than once in the message options.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.4.1
type OptionNotRepeateable struct {
	OptionDef
}

// InvalidOptionValueLength is returned when the length of an option value does not match the expected length.
type InvalidOptionValueLength struct {
	OptionDef
	Length uint16
}

func (e UnmarshalError) Error() string {
	return fmt.Sprintf("parse error at offset %d: %v", e.Offset, e.Cause)
}

func (e UnmarshalError) Unwrap() error {
	return e.Cause
}

func (e UnsupportedVersion) Error() string {
	return fmt.Sprintf("unsupported version %d, expected %d", e.Version, ProtocolVersion)
}

func (e InvalidType) Error() string {
	return fmt.Sprintf("invalid type %s", e.Type)
}

func (e InvalidCode) Error() string {
	return fmt.Sprintf("invalid code %s", e.Code)
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
	if e.Unknown != nil {
		return fmt.Sprintf("invalid option %q value format %q, actual %s", e.Name, e.Unknown, e.Requested)
	}

	return fmt.Sprintf("invalid option %q value format %q, actual %q", e.Name, e.Requested, e.ValueFormat)
}

func (e OptionNotRepeateable) Error() string {
	return fmt.Sprintf("option %q is not repeateable", e.Name)
}
