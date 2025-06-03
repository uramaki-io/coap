package coap

import (
	"strconv"
)

var DefaultSchema = NewSchema()

func init() {
	DefaultSchema.AddOptions(
		IfMatch,
		UriHost,
		ETag,
		IfNoneMatch,
		Observe,
		UriPort,
		LocationPath,
		UriPath,
		ContentFormat,
		MaxAge,
		UriQuery,
		Accept,
		LocationQuery,
		Block1,
		Block2,
		ProxyUri,
		ProxyScheme,
		Size1,
		Size2,
		NoResponse,
	)
}

type Schema struct {
	options map[uint16]OptionDef
}

// Options
var (
	IfMatch       = OptionDef{Code: 1, Name: "IfMatch", ValueFormat: ValueFormatOpaque, Repeatable: true, MaxLen: 8}
	UriHost       = OptionDef{Code: 3, Name: "UriHost", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 255}
	ETag          = OptionDef{Code: 4, Name: "ETag", ValueFormat: ValueFormatOpaque, Repeatable: true, MinLen: 1, MaxLen: 8}
	IfNoneMatch   = OptionDef{Code: 5, Name: "IfNoneMatch", ValueFormat: ValueFormatEmpty}
	Observe       = OptionDef{Code: 6, Name: "Observe", ValueFormat: ValueFormatUint, MaxLen: 3}
	UriPort       = OptionDef{Code: 7, Name: "UriPort", ValueFormat: ValueFormatUint, MaxLen: 2}
	LocationPath  = OptionDef{Code: 8, Name: "LocationPath", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	UriPath       = OptionDef{Code: 11, Name: "UriPath", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	ContentFormat = OptionDef{Code: 12, Name: "ContentFormat", ValueFormat: ValueFormatUint, MaxLen: 2}
	MaxAge        = OptionDef{Code: 14, Name: "MaxAge", ValueFormat: ValueFormatUint, MaxLen: 4}
	UriQuery      = OptionDef{Code: 15, Name: "UriQuery", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	Accept        = OptionDef{Code: 17, Name: "Accept", ValueFormat: ValueFormatUint, MaxLen: 2}
	LocationQuery = OptionDef{Code: 20, Name: "LocationQuery", ValueFormat: ValueFormatString, Repeatable: true, MaxLen: 255}
	Block1        = OptionDef{Code: 27, Name: "Block1", ValueFormat: ValueFormatUint, MaxLen: 3}
	Block2        = OptionDef{Code: 23, Name: "Block2", ValueFormat: ValueFormatUint, MaxLen: 3}
	ProxyUri      = OptionDef{Code: 35, Name: "ProxyUri", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 1034}
	ProxyScheme   = OptionDef{Code: 39, Name: "ProxyScheme", ValueFormat: ValueFormatString, MinLen: 1, MaxLen: 255}
	Size1         = OptionDef{Code: 60, Name: "Size1", ValueFormat: ValueFormatUint, MaxLen: 4}
	Size2         = OptionDef{Code: 28, Name: "Size2", ValueFormat: ValueFormatUint, MaxLen: 4}
	NoResponse    = OptionDef{Code: 258, Name: "NoResponse", ValueFormat: ValueFormatUint, MaxLen: 1}
)

func NewSchema() *Schema {
	return &Schema{
		options: map[uint16]OptionDef{},
	}
}

func (s *Schema) AddOptions(options ...OptionDef) {
	for _, option := range options {
		s.options[option.Code] = option
	}
}

func (s *Schema) Option(code uint16) OptionDef {
	option, ok := s.options[code]
	if !ok {
		return OptionDef{
			Code:        code,
			Name:        strconv.Itoa(int(code)),
			ValueFormat: ValueFormatOpaque,
			MinLen:      0,
			MaxLen:      1034,
		}
	}

	return option
}
