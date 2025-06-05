package coap

import "fmt"

// revive:disable:exported

// Image assets
var (
	MediaTypeImageGIF  = MediaType{Code: 21, Name: `image/gif`}
	MediaTypeImagePNG  = MediaType{Code: 22, Name: `image/png`}
	MediaTypeImageJPEG = MediaType{Code: 23, Name: `image/jpeg`}
)

// Authentication and security
var (
	MediaTypeApplicationCOSEEncrypt0 = MediaType{Code: 16, Name: `application/cose; cose-type="cose-encrypt0"`}
	MediaTypeApplicationCOSEMac0     = MediaType{Code: 17, Name: `application/cose; cose-type="cose-mac0"`}
	MediaTypeApplicationCBORSign1    = MediaType{Code: 18, Name: `application/cbor; cbor-type="cbor-sign1"`}
)

var (
	MediaTypeTextPlain              = MediaType{Code: 0, Name: `text/plain; charset=utf-8`}
	MediaTypeApplicationLinkFormat  = MediaType{Code: 40, Name: `application/link-format`}
	MediaTypeApplicationXML         = MediaType{Code: 41, Name: `application/xml`}
	MediaTypeApplicationOctetStream = MediaType{Code: 42, Name: `application/octet-stream`}
	MediaTypeApplicationExi         = MediaType{Code: 47, Name: `application/exi`}
	MediaTypeApplicationJSON        = MediaType{Code: 50, Name: `application/json`}
	MediaTypeApplicationCBOR        = MediaType{Code: 60, Name: `application/cbor`}
	MediaTypeApplicationCBORSeq     = MediaType{Code: 63, Name: `application/cbor-seq`}
)

// revive:enable:exported

// MediaType indicates payload media type.
type MediaType struct {
	Code uint16
	Name string
}

// UnrecognizedMediaType creates a MediaType instance for unrecognized media types.
func UnrecognizedMediaType(code uint16) MediaType {
	return MediaType{
		Code: code,
	}
}

// Recognized indicates whether the media type is recognized by its name.
func (m MediaType) Recognized() bool {
	return m.Name != ""
}

// String implements fmt.Stringer.
func (m MediaType) String() string {
	if !m.Recognized() {
		return fmt.Sprintf("MediaType(%d)", m.Code)
	}

	return m.Name
}
