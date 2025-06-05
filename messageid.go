package coap

import "sync/atomic"

type MessageID uint16

type MessageIDSource func() MessageID

func MessageIDSequence(start MessageID) MessageIDSource {
	id := atomic.Uint32{}
	id.Store(uint32(start))

	return func() MessageID {
		return MessageID(id.Add(1))
	}
}
