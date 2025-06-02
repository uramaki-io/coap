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
	Length uint8
}

type UnsupportedExtendError struct{}

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
