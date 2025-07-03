package coap

import (
	"fmt"
	"slices"
)

// Response represents a CoAP response message.
type Response struct {
	// Type, defaults to Confirmable.
	Type Type

	// Code
	Code ResponseCode

	// MessageID
	MessageID MessageID

	// Token
	Token Token

	// Options
	Options Options

	// ContentFormat overrides ContentFormat option if set.
	ContentFormat *MediaType

	// LocationPath overrides LocationPath option if not empty.
	LocationPath string

	// LocationQuery overrides LocationQuery options if not empty.
	LocationQuery []string

	// Payload
	Payload []byte
}

// ResponseCode represents a CoAP response message code.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9
type ResponseCode uint8

// Success 2.xx Response Codes
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.1
const (
	Created  ResponseCode = 0x41
	Deleted  ResponseCode = 0x42
	Valid    ResponseCode = 0x43
	Changed  ResponseCode = 0x44
	Content  ResponseCode = 0x45
	Continue ResponseCode = 0x46
)

// Client Error 4.xx Response Codes
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.2
const (
	BadRequest               ResponseCode = 0x80
	Unauthorized             ResponseCode = 0x81
	BadOption                ResponseCode = 0x82
	Forbidden                ResponseCode = 0x83
	NotFound                 ResponseCode = 0x84
	MethodNotAllowed         ResponseCode = 0x85
	NotAcceptable            ResponseCode = 0x86
	Conflict                 ResponseCode = 0x89
	PreconditionFailed       ResponseCode = 0x8c
	RequestEntityTooLarge    ResponseCode = 0x8d
	UnsupportedContentFormat ResponseCode = 0x8f
	RequestEntityIncomplete  ResponseCode = 0x88
	UnprocessableEntity      ResponseCode = 0x96
	TooManyRequests          ResponseCode = 0x9d
)

// Server Error 5.xx Response Codes
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-5.9.3
const (
	InternalServerError  ResponseCode = 0xa0
	NotImplemented       ResponseCode = 0xa1
	BadGateway           ResponseCode = 0xa2
	ServiceUnavailable   ResponseCode = 0xa3
	GatewayTimeout       ResponseCode = 0xa4
	ProxyingNotSupported ResponseCode = 0xa5
	HopLimitReached      ResponseCode = 0xa8
)

func (r *Response) String() string {
	return fmt.Sprintf("Response(Type=%s, MessageID=%d, Code=%s)",
		r.Type,
		r.MessageID,
		r.Code,
	)
}

// AppendBinary appends the binary representation of the Response to the provided data slice.
//
// Returns InvalidType if type is out of range.
//
// Returns InvalidCode if code is not a valid response code.
func (r *Response) AppendBinary(data []byte) ([]byte, error) {
	if r.Type > Reset {
		return data, InvalidType{
			Type: r.Type,
		}
	}

	code := Code(r.Code)
	if code.Class() < 0x01 || code.Class() > 0x10 {
		return data, InvalidCode{
			Code: code,
		}
	}

	options := slices.Clone(r.Options)

	if r.ContentFormat != nil {
		options.SetUint(ContentFormat, uint32(r.ContentFormat.Code))
	}

	if r.LocationPath != "" {
		Must(options.SetAllString(LocationPath, EncodePath(r.LocationPath)))
	}

	if r.LocationQuery != nil {
		Must(options.SetAllString(LocationQuery, slices.Values(r.LocationQuery)))
	}

	msg := Message{
		Header: Header{
			Version: ProtocolVersion,
			Type:    r.Type,
			Code:    code,
			ID:      r.MessageID,
			Token:   r.Token,
		},
		Options: options,
		Payload: r.Payload,
	}

	data, err := msg.AppendBinary(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Decode decodes the Response from the given data using the provided options.
//
// Returns UnmarshalError if the message cannot be decoded.
//
// Returns InvalidCode if the message code class is not in the range of 2.xx to 5.xx.
func (r *Response) Decode(data []byte, opts MarshalOptions) ([]byte, error) {
	if opts.Schema == nil {
		opts.Schema = DefaultSchema
	}

	msg := Message{}

	data, err := msg.Decode(data, opts)
	if err != nil {
		return data, err
	}

	if msg.Code.Class() < 2 || msg.Code.Class() > 5 {
		return data, InvalidCode{
			Code: msg.Code,
		}
	}

	r.Type = msg.Type
	r.Code = ResponseCode(msg.Code)
	r.MessageID = msg.ID
	r.Token = msg.Token
	r.Options = msg.Options
	r.Payload = msg.Payload

	contentFormat, ok := r.Options.Get(ContentFormat)
	if ok {
		code := MustValue(contentFormat.GetUint())
		mediaType := opts.Schema.MediaType(uint16(code))
		r.ContentFormat = &mediaType
	}

	path := MustValue(r.Options.GetAllString(LocationPath))
	r.LocationPath = DecodePath(path)

	query := MustValue(r.Options.GetAllString(LocationQuery))
	r.LocationQuery = slices.Collect(query)

	return data, nil
}

// String implements fmt.Stringer.
func (c ResponseCode) String() string {
	class := (c & 0xe0) >> 5
	detail := c & 0x1f

	return fmt.Sprintf("%d.%02d", class, detail)
}
