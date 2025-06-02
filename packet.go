package coap

type Packet struct {
	Header
	Options

	Token   uint64
	Payload []byte
}
