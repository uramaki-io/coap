package coap

import "testing"

func TestMediaType_RecognizedAndString(t *testing.T) {
	tests := []struct {
		name       string
		mediaType  MediaType
		recognized bool
		str        string
	}{
		{
			name:       "recognized media type",
			mediaType:  MediaTypeApplicationJSON,
			recognized: true,
			str:        "application/json",
		},
		{
			name:       "unrecognized media type",
			mediaType:  UnrecognizedMediaType(999),
			recognized: false,
			str:        "MediaType(999)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.mediaType.Recognized(); got != test.recognized {
				t.Errorf("Recognized() = %v, want %v", got, test.recognized)
			}
			if got := test.mediaType.String(); got != test.str {
				t.Errorf("String() = %q, want %q", got, test.str)
			}
		})
	}
}
