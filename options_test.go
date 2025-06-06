package coap

import (
	"bytes"
	"slices"
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

	t.Run("option not found", func(t *testing.T) {
		opt := URIHost
		opts := Options{}

		_, ok := opts.GetValue(opt)
		if ok {
			t.Fatal("expected option to not exist")
		}

		ok = opts.Contains(opt)
		if ok {
			t.Fatal("expected option to not exist")
		}

		count := opts.Clear(opt)
		if count != 0 {
			t.Fatal("expected clear count to be 0")
		}

		expected := OptionNotFound{
			OptionDef: opt,
		}

		_, err := opts.GetString(opt)
		expectErr(t, err, expected)

		_, err = opts.GetUint(opt)
		expectErr(t, err, expected)

		_, err = opts.GetOpaque(opt)
		expectErr(t, err, expected)
	})
}

func TestOptionsGetSetAll(t *testing.T) {
	uintValues := []uint32{1, 2}
	stringValues := []string{"test", "path"}
	opaqueValues := [][]byte{
		{0x42},
		{0x43, 0x44},
	}

	tests := []struct {
		name   string
		option OptionDef
		set    func(opts *Options, option OptionDef) error
		get    func(opts *Options, option OptionDef) (bool, error)
	}{
		{
			name: "uint",
			option: OptionDef{
				Name:        "UintOption",
				ValueFormat: ValueFormatUint,
				Repeatable:  true,
				MaxLen:      4,
			},
			set: func(opts *Options, option OptionDef) error {
				return opts.SetAllUint(option, slices.Values(uintValues))
			},
			get: func(opts *Options, option OptionDef) (bool, error) {
				seq, err := opts.GetAllUint(option)
				values := slices.Collect(seq)
				equal := slices.Equal(values, uintValues)
				return equal, err
			},
		},
		{
			name: "string",
			option: OptionDef{
				Name:        "StringOption",
				ValueFormat: ValueFormatString,
				Repeatable:  true,
				MaxLen:      1034,
			},
			set: func(opts *Options, option OptionDef) error {
				return opts.SetAllString(option, slices.Values(stringValues))
			},
			get: func(opts *Options, option OptionDef) (bool, error) {
				seq, err := opts.GetAllString(option)
				values := slices.Collect(seq)
				equal := slices.Equal(values, stringValues)
				return equal, err
			},
		},
		{
			name: "opaque",
			option: OptionDef{
				Name:        "OpaqueOption",
				ValueFormat: ValueFormatOpaque,
				Repeatable:  true,
				MaxLen:      1034,
			},
			set: func(opts *Options, option OptionDef) error {
				return opts.SetAllOpaque(option, slices.Values(opaqueValues))
			},
			get: func(opts *Options, option OptionDef) (bool, error) {
				seq, err := opts.GetAllOpaque(option)
				values := slices.Collect(seq)
				equal := slices.EqualFunc(values, opaqueValues, func(a, b []byte) bool {
					return bytes.Equal(a, b)
				})
				return equal, err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := Options{}
			err := test.set(&opts, test.option)
			if err != nil {
				t.Fatal("set all:", err)
			}

			equal, err := test.get(&opts, test.option)
			if err != nil {
				t.Error("get all:", err)
			}

			if !equal {
				t.Error("expected values to be equal")
			}

			if test.option.ValueFormat != ValueFormatUint {
				expected := InvalidOptionValueFormat{
					OptionDef: test.option,
					Requested: ValueFormatUint,
				}

				_, err := opts.GetAllUint(test.option)
				expectErr(t, err, expected)

				err = opts.SetAllUint(test.option, slices.Values(uintValues))
				expectErr(t, err, expected)
			}

			if test.option.ValueFormat != ValueFormatString {
				expected := InvalidOptionValueFormat{
					OptionDef: test.option,
					Requested: ValueFormatString,
				}

				_, err := opts.GetAllString(test.option)
				expectErr(t, err, expected)

				err = opts.SetAllString(test.option, slices.Values(stringValues))
				expectErr(t, err, expected)
			}

			if test.option.ValueFormat != ValueFormatOpaque {
				expected := InvalidOptionValueFormat{
					OptionDef: test.option,
					Requested: ValueFormatOpaque,
				}

				_, err := opts.GetAllOpaque(test.option)
				expectErr(t, err, expected)

				err = opts.SetAllOpaque(test.option, slices.Values(opaqueValues))
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
