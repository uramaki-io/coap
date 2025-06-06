package coap

import (
	"reflect"
	"testing"
)

func TestErrorStringMethods(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "UnmarshalError",
			err: UnmarshalError{
				Offset: 10,
				Cause: TruncatedError{
					Expected: 5,
				},
			},
			want: "unmarshal error at offset 10: truncated input, expected 5 bytes",
		},
		{
			name: "UnsupportedVersion",
			err: UnsupportedVersion{
				Version: 2,
			},
			want: "unsupported version 2, expected 1",
		},
		{
			name: "InvalidType",
			err: InvalidType{
				Type: 3,
			},
			want: "invalid type RST",
		},
		{
			name: "InvalidCode",
			err: InvalidCode{
				Code: 5,
			},
			want: "invalid code 0.05",
		},
		{
			name: "UnsupportedTokenLength",
			err: UnsupportedTokenLength{
				Length: 9,
			},
			want: "unsupported token length 9, max is 8",
		},
		{
			name: "UnsupportedExtendError",
			err:  UnsupportedExtendError{},
			want: "unsupported extend value",
		},
		{
			name: "TruncatedError",
			err: TruncatedError{
				Expected: 8,
			},
			want: "truncated input, expected 8 bytes",
		},
		{
			name: "OptionNotFound",
			err: OptionNotFound{
				OptionDef: OptionDef{
					Name: "Uri-Host",
				},
			},
			want: `option "Uri-Host" not found`,
		},
		{
			name: "InvalidOptionValueLength",
			err: InvalidOptionValueLength{
				OptionDef: OptionDef{
					Name:   "Uri-Host",
					MinLen: 1,
					MaxLen: 10,
				},
				Length: 12,
			},
			want: `expected option "Uri-Host" value length between 1 and 10, got 12`,
		},
		{
			name: "InvalidOptionValueFormat with Unknown",
			err: InvalidOptionValueFormat{
				OptionDef: OptionDef{Name: "Uri-Host"},
				Requested: ValueFormatString,
				Unknown:   reflect.TypeOf(42),
			},
			want: `invalid option "Uri-Host" value format "int", actual string`,
		},
		{
			name: "InvalidOptionValueFormat without Unknown",
			err: InvalidOptionValueFormat{
				OptionDef: OptionDef{
					Name:        "Uri-Host",
					ValueFormat: ValueFormatUint,
				},
				Requested: ValueFormatString,
			},
			want: `invalid option "Uri-Host" value format "string", actual "uint"`,
		},
		{
			name: "OptionNotRepeateable",
			err: OptionNotRepeateable{
				OptionDef: OptionDef{
					Name: "Uri-Host",
				},
			},
			want: `option "Uri-Host" is not repeateable`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
