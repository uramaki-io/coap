package coap

type RequestCode uint8

// Method 0.xx Codes
// https://datatracker.ietf.org/doc/html/rfc7252#section-12.1.1
const (
	Get    RequestCode = 0x01
	Post   RequestCode = 0x02
	Put    RequestCode = 0x03
	Delete RequestCode = 0x04
	Fetch  RequestCode = 0x05
	Patch  RequestCode = 0x06
	IPatch RequestCode = 0x07
)
