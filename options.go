package coap

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
)

// Options represents a collection of CoAP options.
type Options struct {
	data []Option
}

// MakeOptions creates a new Options instance with the provided options.
//
// The options are sorted by their code.
func MakeOptions(data ...Option) Options {
	slices.SortFunc(data, func(a, b Option) int {
		return cmp.Compare(a.Code, b.Code)
	})

	return Options{
		data: data,
	}
}

// Clone creates a shallow copy of the Options.
//
// Opaque values are not copied, so changes to the opaque values in the cloned Options will affect the original Options.
func (o Options) Clone() Options {
	return Options{
		data: slices.Clone(o.data),
	}
}

// Contains checks if the given option is present.
func (o Options) Contains(def OptionDef) bool {
	i := Index(o.data, def)
	return i != -1
}

// Get retrieves the first option matching the definition.
func (o Options) Get(def OptionDef) (Option, bool) {
	i := Index(o.data, def)
	if i == -1 {
		return Option{}, false
	}

	return o.data[i], true
}

// Set creates or updates an option.
func (o *Options) Set(opt Option) {
	i := Index(o.data, opt.OptionDef)
	if i == -1 {
		o.data = append(o.data, opt)
		return
	}

	o.data[i] = opt
}

// GetAll retrieves all options matching the definition.
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

// ClearOption removes all occurrences of the option with matching code.
func (o *Options) ClearOption(def OptionDef) {
	o.data = slices.DeleteFunc(o.data, func(opt Option) bool {
		return opt.Code == def.Code
	})
}

// ClearOptions removes all options.
func (o *Options) ClearOptions() {
	o.data = o.data[0:]
}

// GetValue retrieves the value of the first option matching the definition.
//
// Prefer to use specific methods GetUint, GetOpaque or GetString to ensure type safety and avoid reflect overhead.
func (o Options) GetValue(def OptionDef) (any, bool) {
	opt, ok := o.Get(def)
	if !ok {
		return nil, false
	}

	return opt.GetValue(), true
}

// SetValue creates or updates an option with the given value.
//
// Prefer to use specific methods SetUint, SetOpaque, SetString to ensure type safety and avoid reflect overhead.
func (o *Options) SetValue(def OptionDef, value any) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetValue(value)
	if err != nil {
		return err
	}

	o.Set(opt)

	return nil
}

// GetUint retrieves the value of the first option matching the definition as uint32.
//
// Returns OptionNotFound if the option is not present
// Returns InvalidOptionValueFormat if the option value format is not ValueFormatUint.
func (o Options) GetUint(def OptionDef) (uint32, error) {
	opt, ok := o.Get(def)
	if !ok {
		return 0, OptionNotFound{
			OptionDef: def,
		}
	}

	return opt.GetUint()
}

// SetUint creates or updates an option with the given value as uint32.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatUint.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
func (o *Options) SetUint(def OptionDef, value uint32) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetUint(value)
	if err != nil {
		return err
	}

	o.Set(opt)

	return nil
}

// GetAllUint retrieves all options matching the definition as a sequence of uint32 values.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatUint.
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

// SetAllUint creates or updates all options matching the definition with the given sequence of uint32 values.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatUint.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
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

// GetOpaque retrieves the value of the first option matching the definition as []byte.
//
// Returns OptionNotFound if the option is not present
// Returns InvalidOptionValueFormat if the value format is not ValueFormatOpaque.
func (o Options) GetOpaque(def OptionDef) ([]byte, error) {
	opt, ok := o.Get(def)
	if !ok {
		return nil, OptionNotFound{
			OptionDef: def,
		}
	}

	return opt.GetOpaque()
}

// SetOpaque creates or updates an option with the given value as []byte.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatOpaque.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
func (o *Options) SetOpaque(def OptionDef, value []byte) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetOpaque(value)
	if err != nil {
		return err
	}

	o.Set(opt)

	return nil
}

// GetAllOpaque retrieves all options matching the definition as a sequence of []byte values.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatOpaque.
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

// SetAllOpaque creates or updates all options matching the definition with the given sequence of []byte values.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatOpaque.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
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

// GetString retrieves the value of the first option matching the definition as string.
//
// Returns OptionNotFound if the option is not present
// Returns InvalidOptionValueFormat if the value format is not ValueFormatString.
func (o Options) GetString(def OptionDef) (string, error) {
	opt, ok := o.Get(def)
	if !ok {
		return "", OptionNotFound{
			OptionDef: def,
		}
	}

	return opt.GetString()
}

// SetString creates or updates an option with the given value as string.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatString.
func (o *Options) SetString(def OptionDef, value string) error {
	opt := Option{
		OptionDef: def,
	}

	err := opt.SetString(value)
	if err != nil {
		return err
	}

	o.Set(opt)

	return nil
}

// GetAllString retrieves all options matching the definition as a sequence of string values.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatString.
// Returns InvalidOptionValueLength if the value length does not match the expected length.
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

// SetAllString creates or updates all options matching the definition with the given sequence of string values.
//
// Returns InvalidOptionValueFormat if the value format is not ValueFormatString.
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
//
// It encodes options into the data slice, sorting them by code before encoding.
func (o Options) AppendBinary(data []byte) ([]byte, error) {
	if len(o.data) == 0 {
		return data, nil // no options to encode
	}

	options := slices.Clone(o.data)
	slices.SortFunc(options, func(l, r Option) int {
		return cmp.Compare(l.Code, r.Code)
	})

	prev := uint16(0)
	for _, opt := range options {
		var err error
		data, err = opt.Encode(data, prev)
		if err != nil {
			return data, fmt.Errorf("encode option %q: %w", opt.OptionDef.Name, err)
		}

		prev = opt.Code
	}

	return data, nil
}

// Decode deserializes options from data using schema.
//
// If schema is nil, it panics.
// Returns the remaining data after options have been decoded.
//
// Multiple occurrences of non-repeatable options are treated as unrecognized options.
// Unrecognized options are silently ignored if they are elective.
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

		length := opt.GetLength()
		if length < def.MinLen || length > def.MaxLen {
			return InvalidOptionValueLength{
				OptionDef: def,
				Length:    length,
			}
		}

		i += loc
		o.data[i] = opt
	}

	o.data = slices.AppendSeq(o.data, options)

	return nil
}

// Index returns index of first option with matching code in the options slice returning -1 if not found.
func Index(options []Option, def OptionDef) int {
	return slices.IndexFunc(options, func(opt Option) bool {
		return opt.Code == def.Code
	})
}
