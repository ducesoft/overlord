package proto

import (
	"fmt"
	"sync"
	"time"

	"github.com/ducesoft/overlord/pkg/types"
)

var (
	defaultTime = time.Now()
)

var msgPool = &sync.Pool{
	New: func() interface{} {
		return &Message{}
	},
}

// GetMsgs alloc a slice to the message
func GetMsgs(n int, caps ...int) []*Message {
	largs := len(caps)
	if largs > 1 {
		panic(fmt.Sprintf("optional argument except 1, but get %d", largs))
	}
	var msgs []*Message
	if largs == 0 {
		msgs = make([]*Message, n)
	} else {
		msgs = make([]*Message, n, caps[0])
	}
	for idx := range msgs {
		msgs[idx] = getMsg()
	}
	return msgs
}

// PutMsgs Release message.
func PutMsgs(msgs []*Message) {
	for _, m := range msgs {
		for _, sm := range m.subs {
			sm.clear()
			putMsg(sm)
		}
		for _, r := range m.req {
			r.Put()
		}
		m.clear()
		putMsg(m)
	}
}

// getMsg get the msg from pool
func getMsg() *Message {
	return msgPool.Get().(*Message)
}

// putMsg put the msg into pool
func putMsg(m *Message) {
	msgPool.Put(m)
}

// Message read from client.
type Message struct {
	Type types.CacheType

	req    []Request
	reqNum int
	subs   []*Message
	wg     *sync.WaitGroup

	// Start Time, Write Time, ReadTime, EndTime, Start Pipe Time, End Pipe Time, Start Pipe Time, End Pipe Time
	st, wt, rt, et, spt, ept, sit, eit time.Time
	addr                               string
	err                                error
}

// NewMessage will create new message object.
// this will be used be sub msg req.
func NewMessage() *Message {
	return getMsg()
}

// Reset will clean the msg
func (m *Message) Reset() {
	m.Type = types.CacheTypeUnknown
	m.reqNum = 0
	m.st, m.wt, m.rt, m.et, m.spt, m.ept, m.sit, m.eit = defaultTime, defaultTime, defaultTime, defaultTime, defaultTime, defaultTime, defaultTime, defaultTime
	m.err = nil
}

// clear will clean the msg
func (m *Message) clear() {
	m.Reset()
	m.req = nil
	m.wg = nil
	m.subs = nil
}

// TotalDur will return the total duration of a command.
func (m *Message) TotalDur() time.Duration {
	return m.et.Sub(m.st)
}

// RemoteDur will return the remote execute time of remote mc node.
func (m *Message) RemoteDur() time.Duration {
	return m.rt.Sub(m.wt)
}

// WaitWriteDur ...
func (m *Message) WaitWriteDur() time.Duration {
	return m.wt.Sub(m.st)
}

// PreEndDur ...
func (m *Message) PreEndDur() time.Duration {
	return m.et.Sub(m.rt)
}

// PipeDur ...
func (m *Message) PipeDur() time.Duration {
	return m.ept.Sub(m.spt)
}

// InputDur ...
func (m *Message) InputDur() time.Duration {
	return m.eit.Sub(m.sit)
}

// Addr ...
func (m *Message) Addr() string {
	return m.addr
}

// MarkStart will set the start time of the command to now.
func (m *Message) MarkStart() {
	m.st = time.Now()
}

// MarkWrite will set the write time of the command to now.
func (m *Message) MarkWrite() {
	m.wt = time.Now()
}

// MarkRead will set the read time of the command to now.
func (m *Message) MarkRead() {
	m.rt = time.Now()
}

// MarkEnd will set the end time of the command to now.
func (m *Message) MarkEnd() {
	m.et = time.Now()
}

// MarkStartPipe ...
func (m *Message) MarkStartPipe() {
	m.spt = time.Now()
}

// MarkStartPipe ...
func (m *Message) MarkEndPipe() {
	m.ept = time.Now()
}

// MarkStartInput ...
func (m *Message) MarkStartInput() {
	m.sit = time.Now()
}

// MarkEndInput ...
func (m *Message) MarkEndInput() {
	m.eit = time.Now()
}

// MarkAddr ...
func (m *Message) MarkAddr(addr string) {
	m.addr = addr
}

// ResetSubs will return the Msg data to flush and reset
func (m *Message) ResetSubs() {
	if !m.IsBatch() {
		return
	}
	for i := range m.subs[:m.reqNum] {
		m.subs[i].Reset()
	}
	m.reqNum = 0
}

// NextReq will iterator itself until nil.
func (m *Message) NextReq() (req Request) {
	if m.reqNum < len(m.req) {
		req = m.req[m.reqNum]
		m.reqNum++
	}
	return
}

// WithRequest with proto request.
func (m *Message) WithRequest(req Request) {
	m.req = append(m.req, req)
	m.reqNum++
}

func (m *Message) setRequest(req Request) {
	m.req = m.req[:0]
	m.reqNum = 0
	m.WithRequest(req)
}

// Request returns proto Msg.
func (m *Message) Request() Request {
	if m.req != nil && len(m.req) > 0 {
		return m.req[0]
	}
	return nil
}

// Requests return all request.
func (m *Message) Requests() []Request {
	if m.reqNum == 0 {
		return nil
	}
	return m.req[:m.reqNum]
}

// IsBatch returns whether or not batch.
func (m *Message) IsBatch() bool {
	return m.reqNum > 1
}

// Batch returns sub Msg if is batch.
func (m *Message) Batch() []*Message {
	slen := m.reqNum
	if slen == 0 {
		return nil
	}
	var min = minInt(len(m.subs), slen)
	for i := 0; i < min; i++ {
		m.subs[i].Type = m.Type
		m.subs[i].setRequest(m.req[i])
	}
	delta := slen - len(m.subs)
	for i := 0; i < delta; i++ {
		msg := getMsg()
		msg.Type = m.Type
		msg.st = m.st
		msg.setRequest(m.req[min+i])
		msg.WithWaitGroup(m.wg)
		m.subs = append(m.subs, msg)
	}
	return m.subs[:slen]
}

// WithWaitGroup with wait group.
func (m *Message) WithWaitGroup(wg *sync.WaitGroup) {
	m.wg = wg
}

// Add add wait group.
func (m *Message) Add() {
	if m.wg != nil {
		m.wg.Add(1)
	}
}

// Done mark handle message done.
func (m *Message) Done() {
	if m.wg != nil {
		m.wg.Done()
	}
}

// WithError with error.
func (m *Message) WithError(err error) {
	m.err = err
}

// Err returns error.
func (m *Message) Err() error {
	if m.err != nil {
		return m.err
	}
	if !m.IsBatch() {
		return nil
	}
	for _, s := range m.subs[:m.reqNum] {
		if s.err != nil {
			return s.err
		}
	}
	return nil
}

// ErrMessage return err Msg.
func ErrMessage(err error) *Message {
	return &Message{err: err}
}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Slowlog impl Slowlogger
func (m *Message) Slowlog() (slog *SlowlogEntry) {
	if m.IsBatch() {
		slog = NewSlowlogEntry(m.Type)
		slog.Subs = make([]*SlowlogEntry, m.reqNum)
		for i, req := range m.Requests() {
			slog.Subs[i] = req.Slowlog()
			slog.Subs[i].StartTime = m.st
			slog.Subs[i].RemoteDur = m.subs[i].rt.Sub(m.subs[i].wt)
			slog.Subs[i].TotalDur = m.et.Sub(m.st)
			slog.Subs[i].WaitWriteDur = m.subs[i].WaitWriteDur()
			slog.Subs[i].PreEndDur = m.subs[i].PreEndDur()
			slog.Subs[i].PipeDur = m.subs[i].PipeDur()
			slog.Subs[i].InputDur = m.subs[i].InputDur()
			slog.Subs[i].Addr = m.subs[i].Addr()
		}
	} else {
		slog = m.Request().Slowlog()
		slog.StartTime = m.st
		slog.TotalDur = m.TotalDur()
		slog.RemoteDur = m.RemoteDur()
		slog.WaitWriteDur = m.WaitWriteDur()
		slog.PreEndDur = m.PreEndDur()
		slog.PipeDur = m.PipeDur()
		slog.InputDur = m.InputDur()
		slog.Addr = m.Addr()
	}
	return
}
