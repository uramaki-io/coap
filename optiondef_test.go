package coap

import "testing"

func TestOptionDefMethods(t *testing.T) {
	tests := []struct {
		name       string
		def        OptionDef
		recognized bool
		critical   bool
		unsafe     bool
		noCacheKey bool
		str        string
	}{
		{
			name:       "critical",
			def:        IfMatch,
			recognized: true,
			critical:   true,
			str:        "Option(Name=IfMatch, Code=1, ValueFormat=opaque, MinLen=0, MaxLen=8)",
		},
		{
			name:       "no-cache-key",
			def:        Size1,
			recognized: true,
			noCacheKey: true,
			str:        "Option(Name=Size1, Code=60, ValueFormat=uint, MinLen=0, MaxLen=4)",
		},
		{
			name:   "unrecognized, unsafe",
			def:    UnrecognizedOptionDef(MaxAge.Code, MaxOptionLength),
			unsafe: true,
			str:    "Option(Code=14, ValueFormat=opaque, MaxLen=1024)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.def.Recognized(); got != test.recognized {
				t.Errorf("Recognized() = %v, want %v", got, test.recognized)
			}

			if got := test.def.Critical(); got != test.critical {
				t.Errorf("Critical() = %v, want %v", got, test.critical)
			}

			if got := test.def.Unsafe(); got != test.unsafe {
				t.Errorf("Unsafe() = %v, want %v", got, test.unsafe)
			}

			if got := test.def.NoCacheKey(); got != test.noCacheKey {
				t.Errorf("NoCacheKey() = %v, want %v", got, test.noCacheKey)
			}

			if got := test.def.String(); got != test.str {
				t.Errorf("String() = %q, want %q", got, test.str)
			}
		})
	}
}
