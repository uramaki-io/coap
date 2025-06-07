package coap

import "fmt"

// revive:disable:exported

var (
	IfMatch       = OptionDef{Code: 1, Name: "IfMatch", ValueFormat: ValueFormatOpaque, Repeatable: true, MaxLen: 8}
	URIHost       = OptionDef{Code: 3, Name: "URIHost", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 255}
	ETag          = OptionDef{Code: 4, Name: "ETag", ValueFormat: ValueFormatOpaque, Repeatable: true, MinLen: 1, MaxLen: 8}
	IfNoneMatch   = OptionDef{Code: 5, Name: "IfNoneMatch", ValueFormat: ValueFormatEmpty}
	Observe       = OptionDef{Code: 6, Name: "Observe", ValueFormat: ValueFormatUint, MaxLen: 3}
	URIPort       = OptionDef{Code: 7, Name: "URIPort", ValueFormat: ValueFormatUint, MaxLen: 2}
	LocationPath  = OptionDef{Code: 8, Name: "LocationPath", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	URIPath       = OptionDef{Code: 11, Name: "URIPath", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	ContentFormat = OptionDef{Code: 12, Name: "ContentFormat", ValueFormat: ValueFormatUint, MaxLen: 2}
	MaxAge        = OptionDef{Code: 14, Name: "MaxAge", ValueFormat: ValueFormatUint, MaxLen: 4}
	URIQuery      = OptionDef{Code: 15, Name: "URIQuery", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	Accept        = OptionDef{Code: 17, Name: "Accept", ValueFormat: ValueFormatUint, MaxLen: 2}
	LocationQuery = OptionDef{Code: 20, Name: "LocationQuery", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	Block1        = OptionDef{Code: 27, Name: "Block1", ValueFormat: ValueFormatUint, MaxLen: 3}
	Block2        = OptionDef{Code: 23, Name: "Block2", ValueFormat: ValueFormatUint, MaxLen: 3}
	ProxyURI      = OptionDef{Code: 35, Name: "ProxyURI", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 1034}
	ProxyScheme   = OptionDef{Code: 39, Name: "ProxyScheme", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 255}
	Size1         = OptionDef{Code: 60, Name: "Size1", ValueFormat: ValueFormatUint, MaxLen: 4}
	Size2         = OptionDef{Code: 28, Name: "Size2", ValueFormat: ValueFormatUint, MaxLen: 4}
	NoResponse    = OptionDef{Code: 258, Name: "NoResponse", ValueFormat: ValueFormatUint, MaxLen: 1}
)

// revive:enable:exported

// OptionDef defines a CoAP option with its properties.
//
// Option values are validated agains ValueForman and MinLen/MaxLen.
type OptionDef struct {
	Name        string
	Code        uint16
	ValueFormat ValueFormat
	Repeatable  bool
	MinLen      uint16
	MaxLen      uint16
}

// ValueFormat indicates the format of the option value.
type ValueFormat uint8

const (
	// ValueFormatEmpty indicates an option with no value.
	ValueFormatEmpty ValueFormat = 0x00

	// ValueFormatUint indicates an option with a value that is an unsigned integer.
	ValueFormatUint ValueFormat = 0x01

	// ValueFormatOpaque indicates an option with a value that is an opaque byte sequence.
	ValueFormatOpaque ValueFormat = 0x02

	// ValueFormatString indicates an option with a value that is a UTF-8 string.
	ValueFormatString ValueFormat = 0x03
)

// UnrecognizedOptionDef creates an OptionDef for an unrecognized option code.
func UnrecognizedOptionDef(code uint16, maxLen uint16) OptionDef {
	return OptionDef{
		Code:        code,
		ValueFormat: ValueFormatOpaque,
		MaxLen:      maxLen,
	}
}

// Recognized indicates whether the option is recognized by schema.
func (o OptionDef) Recognized() bool {
	return o.Name != ""
}

// Critical returns true if option critical bit is set.
func (o OptionDef) Critical() bool {
	return o.Code&0x01 == 0x01
}

// Unsafe indicate if proxy should understand this option to forward the message.
func (o OptionDef) Unsafe() bool {
	return o.Code&0x02 == 0x02
}

// NoCacheKey returns true if option should not be used as cache key.
func (o OptionDef) NoCacheKey() bool {
	return o.Code&0x1E == 0x1c
}

// String implements fmt.Stringer.
func (o OptionDef) String() string {
	switch {
	case o.Recognized():
		return fmt.Sprintf("Option(Name=%s, Code=%d, ValueFormat=%s, MinLen=%d, MaxLen=%d)", o.Name, o.Code, o.ValueFormat, o.MinLen, o.MaxLen)
	default:
		return fmt.Sprintf("Option(Code=%d, ValueFormat=%s, MaxLen=%d)", o.Code, o.ValueFormat, o.MaxLen)
	}
}

var valueFormatString = map[ValueFormat]string{
	ValueFormatEmpty:  "empty",
	ValueFormatUint:   "uint",
	ValueFormatOpaque: "opaque",
	ValueFormatString: "string",
}

// String implements fmt.Stringer for ValueFormat.
func (f ValueFormat) String() string {
	s, ok := valueFormatString[f]
	if !ok {
		return "unknown"
	}

	return s
}
