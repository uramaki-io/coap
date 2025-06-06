package coap

import (
	"reflect"
	"testing"
)

func TestErrorStringMethods(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{
			err: UnmarshalError{
				Offset: 10,
				Cause: TruncatedError{
					Expected: 5,
				},
			},
			want: "unmarshal error at offset 10: truncated input, expected 5 bytes",
		},
		{
			err: UnsupportedVersion{
				Version: 2,
			},
			want: "unsupported version 2, expected 1",
		},
		{
			err: InvalidType{
				Type: 3,
			},
			want: "invalid type RST",
		},
		{
			err: InvalidCode{
				Code: 5,
			},
			want: "invalid code 0.05",
		},
		{
			err: UnsupportedTokenLength{
				Length: 9,
			},
			want: "unsupported token length 9, max is 8",
		},
		{
			err:  UnsupportedExtendError{},
			want: "unsupported extend value",
		},
		{
			err: MessageTooLong{
				Limit:  1024,
				Length: 2048,
			},
			want: "message too long, max 1024 bytes, got 2048 bytes",
		},
		{
			err: PayloadTooLong{
				Limit:  512,
				Length: 1024,
			},
			want: "payload too long, max 512 bytes, got 1024 bytes",
		},
		{
			err: TooManyOptions{
				Limit:  10,
				Length: 15,
			},
			want: "too many options, max 10, got 15",
		},
		{
			err: TruncatedError{
				Expected: 8,
			},
			want: "truncated input, expected 8 bytes",
		},
		{
			err: OptionNotFound{
				OptionDef: OptionDef{
					Name: "Uri-Host",
				},
			},
			want: `option "Uri-Host" not found`,
		},
		{
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
			err: InvalidOptionValueFormat{
				OptionDef: OptionDef{Name: "Uri-Host"},
				Requested: ValueFormatString,
				Unknown:   reflect.TypeOf(42),
			},
			want: `invalid option "Uri-Host" value format "int", actual string`,
		},
		{
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
			err: OptionNotRepeateable{
				OptionDef: OptionDef{
					Name: "Uri-Host",
				},
			},
			want: `option "Uri-Host" is not repeateable`,
		},
	}

	for _, test := range tests {
		name := reflect.TypeOf(test.err).Name()
		t.Run(name, func(t *testing.T) {
			got := test.err.Error()
			if got != test.want {
				t.Errorf("Error() = %q, want %q", got, test.want)
			}
		})
	}
}
