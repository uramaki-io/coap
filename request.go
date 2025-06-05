package coap

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

type Request struct {
	Type      Type
	Method    Method
	MessageID MessageID
	Token     Token

	Host  string
	Port  uint16
	Path  string
	Query []string

	Options Options
	Payload []byte
}

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

func (r *Request) String() string {
	return fmt.Sprintf("Request(Type=%s, MessageID=%d, Method=%s, Path=%s)", r.Type, r.MessageID, r.Method, r.Path)
}

func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case FETCH:
		return "FETCH"
	case PATCH:
		return "PATCH"
	case IPATCH:
		return "IPATCH"
	default:
		return Code(m).String()
	}
}

func (r *Request) MarshalBinary() ([]byte, error) {
	data, err := r.AppendBinary(nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// AppendBinary implements encoding.BinaryAppender
func (r *Request) AppendBinary(data []byte) ([]byte, error) {
	if r.Type != Confirmable && r.Type != NonConfirmable {
		return data, UnsupportedType{
			Type: r.Type,
		}
	}

	code := Code(r.Method)
	if r.Method == 0 || code.Class() != 0 {
		return data, UnsupportedCode{
			Code: code,
		}
	}

	if r.Host != "" {
		Must(r.Options.SetString(URIHost, r.Host))
	}

	if r.Port != 0 {
		Must(r.Options.SetUint(URIPort, uint32(r.Port)))
	}

	if r.Path != "" {
		Must(r.Options.SetAllString(URIPath, EncodePath(r.Path)))
	}

	if len(r.Query) != 0 {
		Must(r.Options.SetAllString(URIQuery, slices.Values(r.Query)))
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

	return msg.AppendBinary(data)
}

func (r *Request) UnmarshalBinary(data []byte) error {
	_, err := r.Decode(data, DefaultSchema)
	return err
}

func (r *Request) Decode(data []byte, schema *Schema) ([]byte, error) {
	msg := Message{}

	data, err := msg.Decode(data, schema)
	if err != nil {
		return data, err
	}

	if msg.Type != Confirmable && msg.Type != NonConfirmable {
		return data, UnsupportedType{
			Type: r.Type,
		}
	}

	if msg.Code.Class() != 0 {
		return data, UnsupportedCode{
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

func EncodePath(path string) iter.Seq[string] {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}

	return strings.SplitSeq(path, "/")
}
