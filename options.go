package coap

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
)

type Options struct {
	data []Option
}

func MakeOptions(data ...Option) Options {
	slices.SortFunc(data, func(a, b Option) int {
		return cmp.Compare(a.Code, b.Code)
	})

	return Options{
		data: data,
	}
}

func (o Options) Get(def OptionDef) (Option, bool) {
	i, ok := o.index(def.Code)
	if !ok {
		return Option{}, false
	}

	return o.data[i], true
}

func (o Options) Has(def OptionDef) bool {
	_, ok := o.index(def.Code)
	return ok
}

func (o Options) Seq(def OptionDef) iter.Seq[Option] {
	return func(yield func(Option) bool) {
		i, ok := o.index(def.Code)
		if !ok {
			return
		}

		for _, v := range o.data[i:] {
			if v.Code != def.Code {
				return
			}

			if !yield(v) {
				return
			}
		}
	}
}

func (o *Options) ClearOption(def OptionDef) {
	o.reset(def.Code, 0)
}

func (o *Options) SetOption(opt Option) {
	i := o.reset(opt.Code, 1)

	o.data[i] = opt
}

func (o *Options) ClearOptions() {
	o.data = nil
}

func (o Options) GetUint(def OptionDef) (uint32, error) {
	opt, ok := o.Get(def)
	if !ok {
		return 0, OptionNotFound{
			OptionDef: def,
		}
	}

	return opt.GetUint()
}

func (o *Options) SetUint(def OptionDef, value uint32) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetUint(value)
	if err != nil {
		return err
	}

	o.SetOption(opt)

	return nil
}

func (o Options) GetOpaque(def OptionDef) ([]byte, error) {
	opt, ok := o.Get(def)
	if !ok {
		return nil, OptionNotFound{
			OptionDef: def,
		}
	}

	return opt.GetOpaque()
}

func (o *Options) SetOpaque(def OptionDef, value []byte) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetOpaque(value)
	if err != nil {
		return err
	}

	o.SetOption(opt)

	return nil
}

func (o Options) GetString(def OptionDef) (string, error) {
	opt, ok := o.Get(def)
	if !ok {
		return "", OptionNotFound{
			OptionDef: def,
		}
	}

	return opt.GetString()
}

func (o *Options) SetString(def OptionDef, value string) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetString(value)
	if err != nil {
		return err
	}

	o.SetOption(opt)

	return nil
}

func (o Options) UintSeq(def OptionDef) (iter.Seq[uint32], error) {
	if def.ValueFormat != ValueFormatUint {
		return nil, InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatUint,
		}
	}

	return func(yield func(uint32) bool) {
		for opt := range o.Seq(def) {
			if !yield(opt.uintValue) {
				return
			}
		}
	}, nil
}

func (o *Options) SetUintValues(def OptionDef, values ...uint32) error {
	if def.ValueFormat != ValueFormatUint {
		return InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatUint,
		}
	}

	i := o.reset(def.Code, len(values))
	for j := range len(values) {
		opt := Option{
			OptionDef: def,
		}

		err := opt.SetUint(values[j])
		if err != nil {
			return err
		}

		o.data[i+j] = opt
	}

	return nil
}

func (o Options) OpaqueSeq(def OptionDef) (iter.Seq[[]byte], error) {
	if def.ValueFormat != ValueFormatOpaque {
		return nil, InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatOpaque,
		}
	}

	return func(yield func([]byte) bool) {
		for opt := range o.Seq(def) {
			if !yield(opt.opaqueValue) {
				return
			}
		}
	}, nil
}

func (o *Options) SetOpaqueValues(def OptionDef, values ...[]byte) error {
	if def.ValueFormat != ValueFormatOpaque {
		return InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatOpaque,
		}
	}

	i := o.reset(def.Code, len(values))
	for j := range len(values) {
		o.data[i+j] = Option{
			OptionDef:   def,
			opaqueValue: values[j],
		}
	}

	return nil
}

func (o Options) StringSeq(def OptionDef) (iter.Seq[string], error) {
	if def.ValueFormat != ValueFormatString {
		return nil, InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatString,
		}
	}

	return func(yield func(string) bool) {
		for opt := range o.Seq(def) {
			if !yield(opt.stringValue) {
				return
			}
		}
	}, nil
}

func (o *Options) SetStringValues(def OptionDef, values ...string) error {
	if def.ValueFormat != ValueFormatString {
		return InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatString,
		}
	}

	i := o.reset(def.Code, len(values))
	for j := range len(values) {
		o.data[i+j] = Option{
			OptionDef:   def,
			stringValue: values[j],
		}
	}

	return nil
}

// AppendBinary implements encoding.BinaryAppender
func (o Options) AppendBinary(data []byte) ([]byte, error) {
	if len(o.data) == 0 {
		return data, nil // no options to encode
	}

	prev := uint16(0)
	for _, opt := range o.data {
		var err error
		data, err = opt.Encode(data, prev)
		if err != nil {
			return data, fmt.Errorf("encode option %q: %w", opt.OptionDef.Name, err)
		}

		prev = opt.Code
	}

	return data, nil
}

func (o *Options) Decode(data []byte, schema *Schema) ([]byte, error) {
	if schema == nil {
		panic("schema must not be nil")
	}

	prev := uint16(0)
	options := []Option{}
	for len(data) > 0 && data[0] != PayloadMarker {
		var err error
		var option Option
		data, err = option.Decode(data, prev, schema)
		if err != nil {
			return data, err
		}

		options = append(options, option)

		prev = option.Code
	}

	o.data = options
	return data, nil
}

func (o *Options) reset(code uint16, length int) int {
	if length == 0 {
		return len(o.data)
	}

	i, ok := o.index(code)
	switch {
	// option found, trim if necessary
	case ok:
		count := slices.IndexFunc(o.data[i+1:], func(v Option) bool {
			return v.Code != code
		})

		if count > length {
			o.data = slices.Delete(o.data, i, i+count-length)
			return i
		}
	case i == -1:
		i = len(o.data)
	}

	o.data = slices.Insert(o.data, i, make([]Option, length)...)

	return i
}

func (o Options) index(code uint16) (int, bool) {
	i := slices.IndexFunc(o.data, func(opt Option) bool {
		return opt.Code >= code
	})
	ok := i != -1 && o.data[i].Code == code

	return i, ok
}
