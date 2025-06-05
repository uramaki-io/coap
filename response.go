package coap

import "fmt"

// Response represents a CoAP response message.
type Response struct {
	Type      Type
	Status    Status
	MessageID MessageID
	Token     Token

	Options Options
	Payload []byte
}

type Status uint8

// Success 2.xx Response Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.1
const (
	Created  Status = 0x41
	Deleted  Status = 0x42
	Valid    Status = 0x43
	Changed  Status = 0x44
	Content  Status = 0x45
	Continue Status = 0x46
)

func (c Status) String() string {
	class := (c & 0xe0) >> 5
	detail := c & 0x1f

	return fmt.Sprintf("%d.%02d", class, detail)
}

// Client Error 4.xx Response Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.2
const (
	BadRequest               Status = 0x80
	Unauthorized             Status = 0x81
	BadOption                Status = 0x82
	Forbidden                Status = 0x83
	NotFound                 Status = 0x84
	MethodNotAllowed         Status = 0x85
	NotAcceptable            Status = 0x86
	Conflict                 Status = 0x89
	PreconditionFailed       Status = 0x8c
	RequestEntityTooLarge    Status = 0x8d
	UnsupportedContentFormat Status = 0x8f
	RequestEntityIncomplete  Status = 0x88
	UnprocessableEntity      Status = 0x96
	TooManyRequests          Status = 0x9d
)

// Server Error 5.xx Response Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.3
const (
	InternalServerError  Status = 0xa0
	NotImplemented       Status = 0xa1
	BadGateway           Status = 0xa2
	ServiceUnavailable   Status = 0xa3
	GatewayTimeout       Status = 0xa4
	ProxyingNotSupported Status = 0xa5
	HopLimitReached      Status = 0xa8
)

func (r *Response) AppendBinary(data []byte) ([]byte, error) {
	if r.Type > Reset {
		return data, UnsupportedType{
			Type: r.Type,
		}
	}

	code := Code(r.Status)
	if code.Class() < 0x01 || code.Class() > 0x10 {
		return data, UnsupportedCode{
			Code: code,
		}
	}

	msg := Message{
		Header: Header{
			Version:   ProtocolVersion,
			Type:      r.Type,
			Code:      code,
			MessageID: r.MessageID,
			Token:     r.Token,
		},
		Options: r.Options,
		Payload: r.Payload,
	}

	data, err := msg.AppendBinary(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
