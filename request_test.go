package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRequestRoundtrip(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		request *Request
		options Options // used only for unmarshal comparison
	}{
		{
			name: "valid request with GET method",
			data: []byte{
				0x44, 0x01, 0x00, 0x01, 0xD0, 0xE2, 0x4D, 0xAC, // Header
				0x3b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, // URIHost "example.com"
				0x42, 0x16, 0x33, // URIPort 5683
				0x44, 0x74, 0x65, 0x73, 0x74, // URIPath "/test"
				0x43, 0x61, 0x3d, 0x31, // URIQuery "a=1"
			},
			request: &Request{
				Method:    GET,
				MessageID: 1,
				Token:     []byte{0xD0, 0xE2, 0x4D, 0xAC},
				Host:      "example.com",
				Path:      "/test",
				Port:      5683,
				Query: []string{
					"a=1",
				},
			},
			options: Options{
				MustOptionValue(URIHost, "example.com"),
				MustOptionValue(URIPort, uint32(5683)),
				MustOptionValue(URIPath, "test"),
				MustOptionValue(URIQuery, "a=1"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name+"/marshal", func(t *testing.T) {
			data, err := test.request.MarshalBinary()
			if err != nil {
				t.Fatal("marshal:", err)
			}

			diff := cmp.Diff(test.data, data)
			if diff != "" {
				t.Errorf("data mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			req := &Request{}

			err := req.UnmarshalBinary(test.data)
			if err != nil {
				t.Fatal("unmarshal:", err)
			}

			test.request.Options = test.options
			diff := cmp.Diff(test.request, req, EquateOptions())
			if diff != "" {
				t.Errorf("request mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestDecodeError(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		err  error
	}{
		{
			name: "invalid type",
			data: []byte{0x70, 0xa0, 0x00, 0x00},
			err: InvalidType{
				Type: Reset,
			},
		},
		{
			name: "invalid code",
			data: []byte{0x50, 0x41, 0x42, 0x42},
			err: InvalidCode{
				Code: Code(Created),
			},
		},
		{
			name: "truncated request",
			data: []byte{0x44, 0x01, 0x00, 0x01, 0xD0, 0xE2, 0x4D},
			err: UnmarshalError{
				Offset: 4,
				Cause: TruncatedError{
					Expected: 4,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &Request{}
			err := req.UnmarshalBinary(test.data)
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestAppendBinaryError(t *testing.T) {
	tests := []struct {
		name    string
		request *Request
		err     error
	}{
		{
			name: "invalid type",
			request: &Request{
				Type:   Reset, // not Confirmable or NonConfirmable
				Method: GET,
			},
			err: InvalidType{Type: Reset},
		},
		{
			name: "invalid code",
			request: &Request{
				Type:   Confirmable,
				Method: Method(Created), // not a valid request method (should be 0.xx)
			},
			err: InvalidCode{Code: Code(Created)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.request.AppendBinary(nil)
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestString(t *testing.T) {
	req := &Request{
		Type:      Confirmable,
		MessageID: 123,
		Method:    GET,
		Path:      "/foo/bar",
	}
	want := "Request(Type=CON, MessageID=123, Method=GET, Path=/foo/bar)"
	if got := req.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
