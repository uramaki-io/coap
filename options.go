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

func (o Options) GetFirst(def OptionDef) (Option, error) {
	i, ok := o.index(def)
	if !ok {
		return Option{}, OptionNotFound{
			OptionDef: def,
		}
	}

	return o.data[i], nil
}

func (o Options) Has(def OptionDef) bool {
	_, ok := o.index(def)
	return ok
}

func (o Options) Get(def OptionDef) iter.Seq[Option] {
	return func(yield func(Option) bool) {
		i, ok := o.index(def)
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

func (o *Options) AddOption(opt Option) {
	i, _ := o.index(opt.OptionDef)
	switch {
	case i == -1:
		o.data = append(o.data, opt)
	default:
		o.data = slices.Insert(o.data, i, opt)
	}
}

func (o *Options) ClearOption(def OptionDef) {
	o.data = slices.DeleteFunc(o.data, func(opt Option) bool {
		return opt.Code == def.Code
	})
}

func (o *Options) SetOption(opt Option) {
	i, ok := o.index(opt.OptionDef)
	switch {
	// append option
	case i == -1:
		o.data = append(o.data, opt)
		return
	// insert option
	case !ok:
		o.data = slices.Insert(o.data, i, opt)
		return
	}

	o.data[i] = opt

	// delete all further instances of the same option
	if opt.Repeatable {
		count := slices.IndexFunc(o.data[i+1:], func(v Option) bool {
			return v.Code != opt.Code
		})
		o.data = slices.Delete(o.data, i+1, i+count)
	}
}

func (o *Options) ClearOptions() {
	o.data = nil
}

func (o Options) GetUint(def OptionDef) (uint32, error) {
	opt, err := o.GetFirst(def)
	if err != nil {
		return 0, err
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

func (o Options) GetBytes(def OptionDef) ([]byte, error) {
	opt, err := o.GetFirst(def)
	if err != nil {
		return nil, err
	}

	return opt.GetBytes()
}

func (o *Options) SetBytes(def OptionDef, value []byte) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetBytes(value)
	if err != nil {
		return err
	}

	o.SetOption(opt)

	return nil
}

func (o Options) GetString(def OptionDef) (string, error) {
	opt, err := o.GetFirst(def)
	if err != nil {
		return "", err
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

func (o *Options) GetObserve() (uint32, error) {
	option, err := o.GetFirst(Observe)
	if err != nil {
		return 0, err
	}

	return option.GetUint()
}

func (o *Options) SetObserve(value uint32) {
	o.SetOption(Option{
		OptionDef: Observe,
		uintValue: value,
	})
}

func (o Options) GetURIHost() (string, error) {
	opt, err := o.GetFirst(UriHost)
	if err != nil {
		return "", err
	}

	return opt.GetString()
}

func (o *Options) SetURIHost(host string) error {
	o.SetOption(Option{
		OptionDef:   UriHost,
		stringValue: host,
	})

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
	for {
		switch {
		// end of message
		case len(data) == 0:
			o.data = options
			return data, nil
		// end of options marker
		case data[0] == PayloadMarker:
			o.data = options
			return data[1:], nil
		}

		option := Option{}

		var err error
		data, err = option.Decode(data, prev, schema)
		if err != nil {
			return data, err
		}

		options = append(options, option)

		prev = option.Code
	}
}

func (o *Options) index(def OptionDef) (int, bool) {
	i := slices.IndexFunc(o.data, func(opt Option) bool {
		return opt.Code >= def.Code
	})
	ok := i != -1 && o.data[i].Code == def.Code

	return i, ok
}
