package coap

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

// Request represents a CoAP request message.
type Request struct {
	// Type is the message type, either Confirmable or NonConfirmable.
	//
	// If not set, it defaults to Confirmable.
	Type Type

	// Method
	Method Method

	// MessageID
	MessageID MessageID

	// Token
	Token Token

	// Options
	Options Options

	// Host overrides URIHost option if not empty.
	Host string

	// Port overrides URIPort option if not zero.
	Port uint16

	// Path overrides URIPath options if not empty.
	Path string

	// Query overrides URIQuery options if not empty.
	Query []string

	// ContentFormat overrides ContentFormat option.
	ContentFormat *MediaType

	// Payload
	Payload []byte
}

// Method represents a CoAP request method code.
type Method Code

// Method 0.xx Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-12.1.1
const (
	GET    Method = 0x01
	POST   Method = 0x02
	PUT    Method = 0x03
	DELETE Method = 0x04
	FETCH  Method = 0x05
	PATCH  Method = 0x06
	IPATCH Method = 0x07
)

// String implements fmt.Stringer.
func (r *Request) String() string {
	return fmt.Sprintf("Request(Type=%s, MessageID=%d, Method=%s, Path=%s)", r.Type, r.MessageID, r.Method, r.Path)
}

var methodString = map[Method]string{
	GET:    "GET",
	POST:   "POST",
	PUT:    "PUT",
	DELETE: "DELETE",
	FETCH:  "FETCH",
	IPATCH: "IPATCH",
}

// String implements fmt.Stringer for Method.
func (m Method) String() string {
	s, ok := methodString[m]
	if !ok {
		return fmt.Sprintf("Method(%s)", Code(m))
	}

	return s
}

// MarshalBinary implements encoding.BinaryMarshaler
func (r *Request) MarshalBinary() ([]byte, error) {
	data, err := r.AppendBinary(nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// AppendBinary implements encoding.BinaryAppender
//
// Host, Port, Path, and Query are set in final message options.
func (r *Request) AppendBinary(data []byte) ([]byte, error) {
	if r.Type != Confirmable && r.Type != NonConfirmable {
		return data, InvalidType{
			Type: r.Type,
		}
	}

	code := Code(r.Method)
	if r.Method == 0 || code.Class() != 0 {
		return data, InvalidCode{
			Code: code,
		}
	}

	options := slices.Clone(r.Options)

	if r.Host != "" {
		Must(options.SetString(URIHost, r.Host))
	}

	if r.Port != 0 {
		Must(options.SetUint(URIPort, uint32(r.Port)))
	}

	if r.Path != "" {
		Must(options.SetAllString(URIPath, EncodePath(r.Path)))
	}

	if len(r.Query) != 0 {
		Must(options.SetAllString(URIQuery, slices.Values(r.Query)))
	}

	msg := Message{
		Header: Header{
			Version:   ProtocolVersion,
			Type:      r.Type,
			Code:      code,
			MessageID: r.MessageID,
			Token:     r.Token,
		},
		Options: options,
		Payload: r.Payload,
	}

	return msg.AppendBinary(data)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (r *Request) UnmarshalBinary(data []byte) error {
	_, err := r.Decode(data, DecodeOptions{})
	return err
}

// Decode decodes a CoAP request message from the given data using the provided schema.
//
// Returns UnsupportedType error if the message type is not Confirmable or NonConfirmable.
//
// Returns UnsupportedCode error if the message code is not a valid request method (0.xx).
func (r *Request) Decode(data []byte, opts DecodeOptions) ([]byte, error) {
	msg := Message{}

	data, err := msg.Decode(data, opts)
	if err != nil {
		return data, err
	}

	if msg.Type != Confirmable && msg.Type != NonConfirmable {
		return data, InvalidType{
			Type: msg.Type,
		}
	}

	if msg.Code.Class() != 0 {
		return data, InvalidCode{
			Code: msg.Code,
		}
	}

	host, ok := msg.Get(URIHost)
	if ok {
		r.Host = MustValue(host.GetString())
	}

	port, ok := msg.Get(URIPort)
	if ok {
		r.Port = uint16(MustValue(port.GetUint()))
	}

	path := DecodePath(MustValue(msg.GetAllString(URIPath)))
	query := MustValue(msg.GetAllString(URIQuery))

	r.Type = msg.Type
	r.Method = Method(msg.Code)
	r.MessageID = msg.MessageID
	r.Token = msg.Token
	r.Options = msg.Options
	r.Path = path
	r.Query = slices.Collect(query)

	return data, nil
}

// DecodePath decodes a sequence of path segments into a single path string.
func DecodePath(segments iter.Seq[string]) string {
	if segments == nil {
		return "/"
	}

	path := strings.Builder{}
	for segment := range segments {
		path.WriteRune('/')
		path.WriteString(segment)
	}

	return path.String()
}

// EncodePath encodes a path string into a sequence of path segments.
func EncodePath(path string) iter.Seq[string] {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}

	return strings.SplitSeq(path, "/")
}
