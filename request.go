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
		_ = r.Options.SetString(URIHost, r.Host)
	}

	if r.Port != 0 {
		_ = r.Options.SetUint(URIPort, uint32(r.Port))
	}

	if r.Path != "" {
		_ = r.Options.SetStringValues(URIPath, EncodePath(r.Path)...)
	}

	if len(r.Query) != 0 {
		_ = r.Options.SetStringValues(URIQuery, r.Query...)
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

	var err error
	data, err = msg.AppendBinary(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Request) UnmarshalBinary(data []byte) error {
	msg := Message{}

	err := msg.UnmarshalBinary(data)
	if err != nil {
		return err
	}

	if msg.Type != Confirmable && msg.Type != NonConfirmable {
		return UnsupportedType{
			Type: r.Type,
		}
	}

	if msg.Code.Class() != 0 {
		return UnsupportedCode{
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

	path := DecodePath(MustValue(msg.StringSeq(URIPath)))
	query := MustValue(msg.StringSeq(URIQuery))

	r.Type = msg.Type
	r.Method = Method(msg.Code)
	r.MessageID = msg.MessageID
	r.Token = msg.Token
	r.Options = msg.Options
	r.Path = path
	r.Query = slices.Collect(query)

	return nil
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

func EncodePath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}

	segments := strings.Split(path, "/")

	for i, segment := range segments {
		segments[i] = strings.TrimSpace(segment)
	}

	return segments
}
