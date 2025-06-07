package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestResponseRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		data     []byte
		options  Options // used only for unmarshal comparison
	}{
		{
			name: "valid response with Content",
			response: &Response{
				Type:          Acknowledgement,
				Code:          Content,
				MessageID:     1,
				Token:         []byte{0xD0, 0xE2, 0x4D, 0xAC},
				ContentFormat: &MediaTypeApplicationOctetStream,
				LocationPath:  "/loca/test",
				LocationQuery: []string{"a=1"},
			},
			data: []byte{
				0x64, 0x45, 0x00, 0x01, 0xd0, 0xe2, 0x4d, 0xac,
				0x84, 0x6c, 0x6f, 0x63, 0x61, 0x04, 0x74, 0x65, 0x73, 0x74, // LocationPath "loca/test"
				0x41, 0x2a,
				0x83, 0x61, 0x3d, 0x31, // LocationQuery "a=1"
			},
			options: Options{
				MustOptionValue(ContentFormat, uint32(42)),
				MustOptionValue(LocationPath, "loca"),
				MustOptionValue(LocationPath, "test"),
				MustOptionValue(LocationQuery, "a=1"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name+"/marshal", func(t *testing.T) {
			data, err := test.response.AppendBinary(nil)
			if err != nil {
				t.Fatal("marshal:", err)
			}
			diff := cmp.Diff(test.data, data)
			if diff != "" {
				t.Errorf("data mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(test.name+"/unmarshal", func(t *testing.T) {
			resp := &Response{}
			_, err := resp.Decode(test.data, DecodeOptions{})
			if err != nil {
				t.Fatal("unmarshal:", err)
			}
			test.response.Options = test.options
			diff := cmp.Diff(test.response, resp, EquateOptions())
			if diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResponseDecodeError(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		err  error
	}{
		{
			name: "invalid code class (not 2-5)",
			data: []byte{0x60, 0x01, 0x00, 0x01}, // Type=ACK, Code=0.01 (GET)
			err: InvalidCode{
				Code: 0x01,
			},
		},
		{
			name: "truncated response",
			data: []byte{0x60, 0x45, 0x00}, // incomplete header for Content (2.05)
			err: UnmarshalError{
				Offset: 0,
				Cause: TruncatedError{
					Expected: 4,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := &Response{}
			_, err := resp.Decode(test.data, DecodeOptions{})
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResponseAppendBinaryError(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		err      error
	}{
		{
			name: "invalid type",
			response: &Response{
				Type: Type(99), // invalid type
				Code: Content,
			},
			err: InvalidType{Type: Type(99)},
		},
		{
			name: "invalid code",
			response: &Response{
				Type: Confirmable,
				Code: ResponseCode(0x01), // not a valid response code (should be 2.xx-5.xx)
			},
			err: InvalidCode{Code: Code(0x01)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.response.AppendBinary(nil)
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResponseString(t *testing.T) {
	resp := &Response{
		Type:      Acknowledgement,
		MessageID: 42,
		Code:      Content,
	}
	want := "Response(Type=ACK, MessageID=42, Code=2.05)"
	if got := resp.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
