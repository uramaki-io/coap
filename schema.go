package coap

// DefaultSchema defines well-known CoAP options and media types.
//
// https://www.iana.org/assignments/core-parameters/core-parameters.xhtml#content-formats
var DefaultSchema = NewSchema().
	AddOptions(
		IfMatch,
		URIHost,
		URIPort,
		URIPath,
		URIQuery,
		ETag,
		IfNoneMatch,
		Observe,
		LocationPath,
		LocationQuery,
		ContentFormat,
		MaxAge,
		Accept,
		Block1,
		Block2,
		ProxyURI,
		ProxyScheme,
		Size1,
		Size2,
		NoResponse,
	).
	AddMediaTypes(
		MediaTypeTextPlain,
		MediaTypeImageGIF,
		MediaTypeImagePNG,
		MediaTypeImageJPEG,
		MediaTypeApplicationCOSEEncrypt0,
		MediaTypeApplicationCOSEMac0,
		MediaTypeApplicationCBORSign1,
		MediaTypeApplicationLinkFormat,
		MediaTypeApplicationXML,
		MediaTypeApplicationOctetStream,
		MediaTypeApplicationExi,
		MediaTypeApplicationJSON,
		MediaTypeApplicationCBOR,
		MediaTypeApplicationCBORSeq,
	)

// Schema contains defintions of CoAP options and media types.
//
// It provides methods to add and retrieve options and media types by their code.
type Schema struct {
	options    map[uint16]OptionDef
	mediaTypes map[uint16]MediaType
}

func NewSchema() *Schema {
	return &Schema{
		options:    map[uint16]OptionDef{},
		mediaTypes: map[uint16]MediaType{},
	}
}

func (s *Schema) AddOptions(options ...OptionDef) *Schema {
	for _, option := range options {
		s.options[option.Code] = option
	}

	return s
}

func (s *Schema) AddMediaTypes(mediaTypes ...MediaType) *Schema {
	for _, mediaType := range mediaTypes {
		s.mediaTypes[mediaType.Code] = mediaType
	}

	return s
}

func (s *Schema) Option(code uint16) OptionDef {
	option, ok := s.options[code]
	if !ok {
		return UnrecognizedOptionDef(code)
	}

	return option
}

func (s *Schema) MediaType(code uint16) MediaType {
	mediaType, ok := s.mediaTypes[code]
	if !ok {
		return UnrecognizedMediaType(code)
	}

	return mediaType
}
