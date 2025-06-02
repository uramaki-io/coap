package coap

import "fmt"

type ResponseCode uint8

func (c ResponseCode) String() string {
	class := (c & 0xe0) >> 5
	detail := c & 0x1f

	return fmt.Sprintf("%d.%02d", class, detail)
}

const ResponseCodeUnknown ResponseCode = 0xFF

// Success 2.xx Response Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.1
const (
	Created ResponseCode = iota + 0x41
	Deleted
	Valid
	Changed
	Content
	Continue
)

// Client Error 4.xx Response Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.2
const (
	BadRequest ResponseCode = iota + 0x80
	Unauthorized
	BadOption
	Forbidden
	NotFound
	MethodNotAllowed
	NotAcceptable
	Conflict                 = 0x89
	PreconditionFailed       = 0x8c
	RequestEntityTooLarge    = 0x8d
	UnsupportedContentFormat = 0x8f
	RequestEntityIncomplete  = 0x88
	UnprocessableEntity      = 0x96
	TooManyRequests          = 0x9d
)

// Server Error 5.xx Response Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.3
const (
	InternalServerError ResponseCode = iota + 0xa0
	NotImplemented
	BadGateway
	ServiceUnavailable
	GatewayTimeout
	ProxyingNotSupported
	HopLimitReached = 0xa8
)
