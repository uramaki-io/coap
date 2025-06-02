package coap

import (
	"fmt"
	"iter"
	"slices"
)

type Options struct {
	data []Option
}

type OptionNotFound struct {
	OptionDef
}

func (e OptionNotFound) Error() string {
	return fmt.Sprintf("option %q not found", e.Name)
}

func (o Options) GetFirst(def OptionDef) (*Option, error) {
	i := slices.IndexFunc(o.data, func(v Option) bool {
		return v.Code == def.Code
	})
	if i == -1 {
		return nil, OptionNotFound{
			OptionDef: def,
		}
	}

	return &o.data[i], nil
}

func (o Options) Has(def OptionDef) bool {
	return slices.ContainsFunc(o.data, func(v Option) bool {
		return v.Code == def.Code
	})
}

func (o Options) Get(def OptionDef) iter.Seq[Option] {
	return func(yield func(Option) bool) {
		i := slices.IndexFunc(o.data, func(v Option) bool {
			return v.Code == def.Code
		})
		if i == -1 {
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
	i := slices.IndexFunc(o.data, func(v Option) bool {
		return v.Code >= opt.Code
	})
	switch {
	case i == -1:
		o.data = append(o.data, opt)
	default:
		o.data = slices.Insert(o.data, i, opt)
	}
}

func (o *Options) ClearOption(opt Option) {
	o.data = slices.DeleteFunc(o.data, func(v Option) bool {
		return v.Code == opt.Code
	})
}

func (o *Options) SetOption(opt Option) {
	o.ClearOption(opt)
	o.AddOption(opt)
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

func (o Options) SetBytes(def OptionDef, value []byte) error {
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

func (o Options) SetString(def OptionDef, value string) error {
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
