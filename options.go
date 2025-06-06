package coap

import (
	"cmp"
	"iter"
	"slices"
)

// Options represents a collection of CoAP options.
type Options []Option

// SortOptions sorts the options by their code in ascending order.
//
// Returns a new slice of options sorted by code.
func SortOptions(options Options) Options {
	options = slices.Clone(options)
	slices.SortFunc(options, func(l, r Option) int {
		return cmp.Compare(l.Code, r.Code)
	})

	return options
}

// Contains checks if the given option is present.
func (o Options) Contains(def OptionDef) bool {
	i := Index(o, def)
	return i != -1
}

// Get retrieves the first option matching the definition.
func (o Options) Get(def OptionDef) (Option, bool) {
	i := Index(o, def)
	if i == -1 {
		return Option{}, false
	}

	return o[i], true
}

// Set creates or updates an option.
func (o *Options) Set(opt Option) {
	i := Index(*o, opt.OptionDef)
	if i == -1 {
		*o = append(*o, opt)
		return
	}

	(*o)[i] = opt
}

// GetAll retrieves all options matching the definition.
func (o Options) GetAll(def OptionDef) iter.Seq[Option] {
	return func(yield func(Option) bool) {
		for _, v := range o {
			if v.Code != def.Code {
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

// Clear removes all occurrences of the option with matching code.
//
// Returns number of options removed.
func (o *Options) Clear(def OptionDef) int {
	length := len(*o)
	*o = slices.DeleteFunc(*o, func(opt Option) bool {
		return opt.Code == def.Code
	})

	return length - len(*o)
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
// Returns OptionNotFound if the option is not present.
//
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
//
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
//
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
// Returns OptionNotFound if the option is not present.
//
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
//
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
//
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
// Returns OptionNotFound if the option is not present.
//
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
//
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

// Encode encodes options into the data slice.
//
// If there are no options to encode, it returns the data slice unchanged.
func (o Options) Encode(data []byte) []byte {
	if len(o) == 0 {
		return data // no options to encode
	}

	options := SortOptions(o)
	prev := uint16(0)
	for _, opt := range options {
		data = opt.Encode(data, prev)
		prev = opt.Code
	}

	return data
}

// Decode decodes options from data using schema.
//
// Returns the remaining data after options have been decoded.
//
// Returns TruncatedError if the data is too short to decode the option.
//
// Returns InvalidOptionValueLength if the decoded length does not match the expected length defined in OptionDef.
//
// Multiple occurrences of non-repeatable options are treated as unrecognized options.
// Unrecognized options are silently ignored if they are elective.
func (o *Options) Decode(data []byte, schema *Schema) ([]byte, error) {
	if schema == nil {
		schema = DefaultSchema
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

	*o = options
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
		if i == len(*o) {
			break
		}

		loc := Index((*o)[i:], def)
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
		(*o)[i] = opt
	}

	*o = slices.AppendSeq(*o, options)

	return nil
}

// Index returns index of first option with matching code in the options slice returning -1 if not found.
func Index(options []Option, def OptionDef) int {
	return slices.IndexFunc(options, func(opt Option) bool {
		return opt.Code == def.Code
	})
}
