package coap

import (
	"context"
	"math/rand/v2"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ACKTimeout      = 2 * time.Second
	ACKRandomFactor = 1.5
	MaxRetransmit   = 4
)

var NoopRetransmitErrorHandler RetransmitErrorHandler = func(_ *Message, _ error) {}

// Conn represents a CoAP connection over a net.PacketConn with retransmission of Confirmable messages.
type Conn struct {
	delegate net.PacketConn
	opts     ConnOptions

	rx *Reader
	tx *Writer

	closed atomic.Bool
	done   chan struct{}
	add    chan WriteOp
	remove chan MessageID
}

// ConnOptions holds options for creating a new CoAP connection.
type ConnOptions struct {
	RetransmitOptions
	MarshalOptions
}

// RetransmitOptions holds options for reliable message transmission.
type RetransmitOptions struct {
	ACKTimeout      time.Duration
	ACKRandomFactor float64
	MaxRetransmit   uint
	MaxTransmitWait time.Duration
	MaxTransmitSpan time.Duration
	ErrorHandler    RetransmitErrorHandler
}

type RetransmitErrorHandler func(msg *Message, err error)

// Reader reads messages from net.PacketConn using provided MarshalOptions.
type Reader struct {
	conn net.PacketConn
	opts MarshalOptions

	mtx sync.Mutex
	buf []byte
}

// Writer writes messages to net.PacketConn using provided MarshalOptions.
type Writer struct {
	conn net.PacketConn
	opts MarshalOptions

	mtx sync.Mutex
	buf []byte
}

// RetransmitQueue manages retransmission of Confirmable messages until they are acknowledged or the maximum retransmission limit/time is reached.
type RetransmitQueue struct {
	opts RetransmitOptions
	data []WriteOp
}

// WriteOp represents a write operation for a Confirmable message that needs retransmission.
type WriteOp struct {
	Message    *Message
	Addr       net.Addr
	Start      time.Time
	Retransmit uint
	Timeout    time.Duration
	Next       time.Time
}

// ListenPacket instantiates a new Conn that listens for incoming packets on the specified network and address.
func ListenPacket(ctx context.Context, network string, address string, opts ConnOptions) (*Conn, error) {
	cfg := net.ListenConfig{}
	delegate, err := cfg.ListenPacket(ctx, network, address)
	if err != nil {
		return nil, err
	}

	return NewConn(delegate, opts), nil
}

// NewConn instantiates a new Conn with the provided PacketConn and options.
func NewConn(delegate net.PacketConn, opts ConnOptions) *Conn {
	rx := NewReader(delegate, opts.MarshalOptions)
	tx := NewWriter(delegate, opts.MarshalOptions)

	conn := &Conn{
		delegate: delegate,
		opts:     opts,
		rx:       rx,
		tx:       tx,
		add:      make(chan WriteOp, 1),
		remove:   make(chan MessageID, 1),
		done:     make(chan struct{}, 1),
	}

	go conn.run()

	return conn
}

// Close closes the connection and stops the retransmission queue.
func (c *Conn) Close() error {
	if !c.closed.Swap(true) {
		close(c.done)
	}

	return c.delegate.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.delegate.LocalAddr()
}

// Read reads a message from the connection and returns the address it was received from.
func (c *Conn) Read(msg *Message) (addr net.Addr, err error) {
	if c.closed.Load() {
		return nil, net.ErrClosed
	}

	addr, err = c.rx.Read(msg)
	if err != nil {
		return addr, err
	}

	if msg.Type != Acknowledgement && msg.Type != Reset {
		return addr, nil
	}

	select {
	case <-c.done:
		return addr, net.ErrClosed
	case c.remove <- msg.ID:
	}

	return addr, nil
}

// Write sends a message to the specified address and handles retransmission for Confirmable messages.
func (c *Conn) Write(msg *Message, addr net.Addr) error {
	if c.closed.Load() {
		return net.ErrClosed
	}

	err := c.tx.Write(msg, addr)
	if err != nil {
		return err
	}

	if msg.Type != Confirmable {
		return nil
	}

	now := time.Now()
	jitter := rand.N(time.Duration(float64(c.opts.ACKTimeout) * c.opts.ACKRandomFactor))
	timeout := c.opts.ACKTimeout + jitter
	op := WriteOp{
		Message: msg,
		Addr:    addr,
		Start:   now,
		Timeout: timeout,
		Next:    now.Add(timeout),
	}

	select {
	case <-c.done:
		return net.ErrClosed
	case c.add <- op:
		return nil
	}
}

func (c *Conn) run() {
	queue := NewRetransmitQueue(c.opts.RetransmitOptions)
	retransmits := []WriteOp{}

	t := time.NewTimer(c.opts.ACKTimeout)
	defer t.Stop()
	for {
		select {
		case <-c.done:
			queue.Close()
			return
		case op := <-c.add:
			queue.Add(op)
		case id := <-c.remove:
			queue.Remove(id)
		case <-t.C:
			retransmits = queue.Retransmit(time.Now(), retransmits)
			for _, op := range retransmits {
				err := c.tx.Write(op.Message, op.Addr)
				if err != nil {
					queue.opts.ErrorHandler(op.Message, err)
					continue
				}
			}
		}

		t.Reset(queue.Next(time.Now()))
	}
}

// NewReader instantiates a new Reader that can read messages from the specified PacketConn.
func NewReader(conn net.PacketConn, opts MarshalOptions) *Reader {
	return &Reader{
		conn: conn,
		opts: opts,
		buf:  make([]byte, opts.MaxMessageLength),
	}
}

// Read reads a message from the PacketConn and decodes it into the provided Message.
func (r *Reader) Read(msg *Message) (addr net.Addr, err error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.buf = r.buf[:0]
	_, addr, err = r.conn.ReadFrom(r.buf)
	if err != nil {
		return addr, err
	}

	_, err = msg.Decode(r.buf, r.opts)
	return addr, err
}

// NewWriter instantiates a new Writer that can send messages over the specified PacketConn.
func NewWriter(conn net.PacketConn, opts MarshalOptions) *Writer {
	return &Writer{
		conn: conn,
		opts: opts,
		buf:  make([]byte, opts.MaxMessageLength),
	}
}

// Write sends a message to the specified address.
func (w *Writer) Write(msg *Message, addr net.Addr) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.buf = w.buf[:0]
	w.buf = msg.Encode(w.buf)

	_, err := w.conn.WriteTo(w.buf, addr)
	return err
}

// NewRetransmitQueue instantiate a new retransmit queue with the given writer and options.
//
// If ErrorHandler is not set, it defaults to NoopRetransmitErrorHandler.
func NewRetransmitQueue(opts RetransmitOptions) *RetransmitQueue {
	if opts.ErrorHandler == nil {
		opts.ErrorHandler = NoopRetransmitErrorHandler
	}

	return &RetransmitQueue{
		opts: opts,
	}
}

// Add adds op to the retransmit queue.
func (q *RetransmitQueue) Add(op WriteOp) {
	q.data = append(q.data, op)
}

// Remove removes op from the retransmit queue by its message ID.
func (q *RetransmitQueue) Remove(id MessageID) (WriteOp, bool) {
	i := slices.IndexFunc(q.data, func(op WriteOp) bool {
		return op.Message.ID == id
	})
	if i == -1 {
		return WriteOp{}, false
	}

	op := q.data[i]
	q.data = slices.Delete(q.data, i, i+1)

	return op, true
}

// Close clears the retransmit queue and calls the error handler for each message with net.ErrClosed.
func (q *RetransmitQueue) Close() {
	for _, op := range q.data {
		q.opts.ErrorHandler(op.Message, net.ErrClosed)
	}

	q.data = q.data[:0]
}

// Retransmit retransmits messages that are pending acknowledgement.
//
// ErrorHandler is called when message retransmission exceeds limits.
//
// https://datatracker.ietf.org/doc/html/rfc7252#section-4.8.2
func (q *RetransmitQueue) Retransmit(now time.Time, retransmits []WriteOp) []WriteOp {
	retransmits = retransmits[:0]

	i := 0
	for _, op := range q.data {
		switch {
		// noop
		case op.Next.After(now):
			q.data[i] = op
		// MAX_RETRANSMIT is the maximum number of retransmissions of a Confirmable message
		case op.Retransmit == q.opts.MaxRetransmit:
			q.opts.ErrorHandler(op.Message, RetransmitRetryLimit{
				Retransmit:    op.Retransmit,
				MaxRetransmit: q.opts.MaxRetransmit,
			})
			continue
		// MAX_TRANSMIT_WAIT is the maximum time from the first transmission
		// of a Confirmable message to the time when the sender gives up on
		// receiving an acknowledgement or reset
		case op.Start.Add(q.opts.MaxTransmitWait).Before(now):
			q.opts.ErrorHandler(op.Message, RetransmitWaitLimit{
				MaxTransmitWait: q.opts.MaxTransmitWait,
			})
			continue
		// MAX_TRANSMIT_SPAN is the maximum time from the first transmission
		// of a Confirmable message to its last retransmission.
		case op.Start.Add(q.opts.MaxTransmitSpan).Before(now):
			q.data[i] = op
		// op needs retransmit
		default:
			op.Timeout *= 2
			op.Retransmit++
			op.Next = now.Add(op.Timeout)
			q.data[i] = op
			retransmits = append(retransmits, op)
		}

		i++
	}

	// resize after skipping expired ops
	q.data = slices.Delete(q.data, i, len(q.data))

	return retransmits
}

// Next returns the next retransmit time.
func (q *RetransmitQueue) Next(now time.Time) time.Duration {
	next := now.Add(q.opts.ACKTimeout)

	for _, op := range q.data {
		if op.Next.Before(next) {
			next = op.Next
		}
	}

	return next.Sub(now)
}
