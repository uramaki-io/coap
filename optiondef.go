package coap

import "strconv"

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

type OptionDef struct {
	Name        string
	Code        uint16
	ValueFormat ValueFormat
	Repeatable  bool
	MinLen      uint16
	MaxLen      uint16
}

type ValueFormat uint8

const (
	ValueFormatEmpty  ValueFormat = 0x00
	ValueFormatUint   ValueFormat = 0x01
	ValueFormatOpaque ValueFormat = 0x02
	ValueFormatString ValueFormat = 0x03
)

func UnrecognizedOptionDef(code uint16) OptionDef {
	return OptionDef{
		Code:        code,
		ValueFormat: ValueFormatOpaque,
		MaxLen:      1034,
	}
}

func (o OptionDef) Recognized() bool {
	return o.Name != ""
}

// Critical returns true if option critical bit is set.
func (o OptionDef) Critical() bool {
	return o.Code&0x01 == 0x01
}

func (o OptionDef) Unsafe() bool {
	return o.Code&0x02 == 0x02
}

func (o OptionDef) NoCacheKey() bool {
	return o.Code&0x1E == 0x1c
}

func (o OptionDef) String() string {
	return o.Name
}

var valueFormatString = map[ValueFormat]string{
	ValueFormatEmpty:  "empty",
	ValueFormatUint:   "uint",
	ValueFormatOpaque: "opaque",
	ValueFormatString: "string",
}

func (f ValueFormat) String() string {
	s, ok := valueFormatString[f]
	if !ok {
		panic("unknown value format: " + strconv.FormatUint(uint64(f), 10))
	}

	return s
}
