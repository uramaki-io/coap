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
	i := Index(o.data, def)
	if i == -1 {
		return Option{}, false
	}

	return o.data[i], true
}

func (o Options) Has(def OptionDef) bool {
	i := Index(o.data, def)
	return i != -1
}

func (o Options) GetAll(def OptionDef) iter.Seq[Option] {
	return func(yield func(Option) bool) {
		for _, v := range o.data {
			if v.Code != def.Code {
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

func (o *Options) ClearOption(def OptionDef) {
	o.data = slices.DeleteFunc(o.data, func(opt Option) bool {
		return opt.Code == def.Code
	})
}

func (o *Options) SetOption(opt Option) {
	i := Index(o.data, opt.OptionDef)
	if i == -1 {
		o.data = append(o.data, opt)
		return
	}

	o.data[i] = opt
}

func (o *Options) ClearOptions() {
	o.data = o.data[0:]
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

func (o Options) GetAllUint(def OptionDef) (iter.Seq[uint32], error) {
	if def.ValueFormat != ValueFormatUint {
		return nil, InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatUint,
		}
	}

	return func(yield func(uint32) bool) {
		for opt := range o.GetAll(def) {
			if !yield(opt.uintValue) {
				return
			}
		}
	}, nil
}

func (o *Options) SetAllUint(def OptionDef, values iter.Seq[uint32]) error {
	if def.ValueFormat != ValueFormatUint {
		return InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatUint,
		}
	}

	return o.setAll(def, func(yield func(Option) bool) {
		for v := range values {
			opt := Option{
				OptionDef: def,
				uintValue: v,
			}
			if !yield(opt) {
				return
			}
		}
	})
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

func (o Options) GetAllOpaque(def OptionDef) (iter.Seq[[]byte], error) {
	if def.ValueFormat != ValueFormatOpaque {
		return nil, InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatOpaque,
		}
	}

	return func(yield func([]byte) bool) {
		for opt := range o.GetAll(def) {
			if !yield(opt.opaqueValue) {
				return
			}
		}
	}, nil
}

func (o *Options) SetAllOpaque(def OptionDef, values iter.Seq[[]byte]) error {
	if def.ValueFormat != ValueFormatOpaque {
		return InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatOpaque,
		}
	}

	return o.setAll(def, func(yield func(Option) bool) {
		for v := range values {
			opt := Option{
				OptionDef:   def,
				opaqueValue: v,
			}
			if !yield(opt) {
				return
			}
		}
	})
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

func (o Options) GetAllString(def OptionDef) (iter.Seq[string], error) {
	if def.ValueFormat != ValueFormatString {
		return nil, InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatString,
		}
	}

	return func(yield func(string) bool) {
		for opt := range o.GetAll(def) {
			if !yield(opt.stringValue) {
				return
			}
		}
	}, nil
}

func (o *Options) SetAllString(def OptionDef, values iter.Seq[string]) error {
	if def.ValueFormat != ValueFormatString {
		return InvalidOptionValueFormat{
			OptionDef: def,
			Requested: ValueFormatString,
		}
	}

	return o.setAll(def, func(yield func(Option) bool) {
		for v := range values {
			opt := Option{
				OptionDef:   def,
				stringValue: v,
			}
			if !yield(opt) {
				return
			}
		}
	})
}

// AppendBinary implements encoding.BinaryAppender
func (o Options) AppendBinary(data []byte) ([]byte, error) {
	if len(o.data) == 0 {
		return data, nil // no options to encode
	}

	slices.SortFunc(o.data, func(l, r Option) int {
		return cmp.Compare(l.Code, r.Code)
	})

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

		// Each occurence of non-repeatable option has to be treated as unrecognized
		// https://datatracker.ietf.org/doc/html/rfc7252#section-5.4.5
		if !option.Repeatable && option.Code == prev {
			option.OptionDef = UnrecognizedOptionDef(option.Code)
		}

		prev = option.Code

		// Unrecognized elective options MUST be silently ignored
		// https://datatracker.ietf.org/doc/html/rfc7252#section-5.4.1
		if !option.Recognized() && !option.Critical() {
			continue
		}

		options = append(options, option)
	}

	o.data = options
	return data, nil
}

func (o *Options) setAll(def OptionDef, options iter.Seq[Option]) error {
	if !def.Repeatable {
		return OptionNotRepeateable{
			OptionDef: def,
		}
	}

	i := 0
	for opt := range options {
		if i == len(o.data) {
			break
		}

		loc := Index(o.data[i:], def)
		if loc == -1 {
			break
		}

		i += loc
		o.data[i] = opt
	}

	o.data = slices.AppendSeq(o.data, options)

	return nil
}

func Index(options []Option, def OptionDef) int {
	return slices.IndexFunc(options, func(opt Option) bool {
		return opt.Code == def.Code
	})
}
