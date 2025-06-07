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
	}{
		{
			name:       "critical",
			def:        IfMatch,
			recognized: true,
			critical:   true,
		},
		{
			name:       "no-cache-key",
			def:        Size1,
			recognized: true,
			noCacheKey: true,
		},
		{
			name:   "unrecognized, unsafe",
			def:    UnrecognizedOptionDef(MaxAge.Code),
			unsafe: true,
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
		})
	}
}
