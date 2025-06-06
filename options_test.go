package coap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestOptionsGetSet(t *testing.T) {
	tests := []struct {
		name   string
		option OptionDef
		value  any
	}{
		{
			name:   "string option",
			option: URIHost,
			value:  "example.com",
		},
		{
			name:   "uint option",
			option: URIPort,
			value:  uint32(0x4242),
		},
		{
			name:   "opaque option",
			option: IfMatch,
			value:  bytes4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := Options{}
			err := opts.SetValue(test.option, test.value)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			value, ok := opts.GetValue(test.option)
			if !ok {
				t.Fatal("expected value to exist:", err)
			}

			diff := cmp.Diff(test.value, value)
			if diff != "" {
				t.Errorf("value mismatch (-want +got):\n%s", diff)
			}

			if test.option.ValueFormat != ValueFormatUint {
				expected := InvalidOptionValueFormat{
					OptionDef: test.option,
					Requested: ValueFormatUint,
				}

				_, err := opts.GetUint(test.option)
				expectErr(t, err, expected)

				err = opts.SetUint(test.option, 0x4242)
				expectErr(t, err, expected)
			}

			if test.option.ValueFormat != ValueFormatString {
				expected := InvalidOptionValueFormat{
					OptionDef: test.option,
					Requested: ValueFormatString,
				}

				_, err := opts.GetString(test.option)
				expectErr(t, err, expected)

				err = opts.SetString(test.option, "example.com")
				expectErr(t, err, expected)
			}

			if test.option.ValueFormat != ValueFormatOpaque {
				expected := InvalidOptionValueFormat{
					OptionDef: test.option,
					Requested: ValueFormatOpaque,
				}

				_, err := opts.GetOpaque(test.option)
				expectErr(t, err, expected)

				err = opts.SetOpaque(test.option, bytes4)
				expectErr(t, err, expected)
			}
		})
	}
}

func EquateOptions() cmp.Option {
	return cmp.Options{
		cmp.Transformer("Options", func(o Options) []string {
			data := SortOptions(o)
			opts := make([]string, 0, len(data))
			for _, opt := range data {
				opts = append(opts, opt.String())
			}

			return opts
		}),
		cmpopts.IgnoreUnexported(Option{}),
	}
}

func TestOptionsDecodeUnrecognized(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		options []Option
		err     error
	}{
		{
			name:    "ignore unrecognized elective option",
			data:    []byte{0xE0, 0xFE, 0xF1},
			options: []Option{},
		},
		{
			name: "multiple occurences of non-repeatable elective option",
			data: []byte{0x72, 0x42, 0x42, 0x02, 0x42, 0x42},
			options: []Option{
				MustOptionValue(URIPort, uint32(0x4242)),
				MustOptionValue(UnrecognizedOptionDef(URIPort.Code), []byte{}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := Options{}
			_, err := opts.Decode(test.data, nil)
			diff := cmp.Diff(test.err, err, cmpopts.EquateErrors())
			if diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}

			diff = cmp.Diff(test.options, opts, EquateOptions())
			if diff != "" {
				t.Errorf("options mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
