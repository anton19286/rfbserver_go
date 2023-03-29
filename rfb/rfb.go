// rfb.go
package rfb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	v8 = "RFB 003.008\n"
)

const (
	// Client -> Server
	CmdSetPixelFormat           = 0
	CmdSetEncodings             = 2
	CmdFramebufferUpdateRequest = 3
	CmdKeyEvent                 = 4
	CmdPointerEvent             = 5
	CmdClientCutText            = 6
	// Server -> Client
	CmdFramebufferUpdate = 0
)

type messageInfo struct {
	Name      string
	HeaderLen int
	LenCalc   func([]byte) int
}

var infoMap map[byte]messageInfo

func init() {
	infoMap = make(map[byte]messageInfo)
	infoMap[CmdSetPixelFormat] = messageInfo{
		"CmdSetPixelFormat", 3,
		func(b []byte) int { return 16 },
	}
	infoMap[CmdSetEncodings] = messageInfo{
		"CmdSetEncodings", 3,
		func(b []byte) int { return int(4 * binary.BigEndian.Uint16(b[1:])) },
	}
	infoMap[CmdFramebufferUpdateRequest] = messageInfo{
		"CmdFramebufferUpdateRequest", 0,
		func(b []byte) int { return 9 },
	}
	infoMap[CmdKeyEvent] = messageInfo{
		"CmdKeyEvent", 0,
		func(b []byte) int { return 7 },
	}
	infoMap[CmdPointerEvent] = messageInfo{
		"CmdPointerEvent", 0,
		func(b []byte) int { return 5 },
	}
	infoMap[CmdClientCutText] = messageInfo{
		"CmdClientCutText", 7,
		func(b []byte) int { return int(binary.BigEndian.Uint32(b[3:])) },
	}
}

const maxHeaderLen = 16

// A Conn reads and writes RFB messages.
type Conn struct {
	rwc 	io.ReadWriteCloser
	rw      io.ReadWriter
	header  [maxHeaderLen]byte
	rbuf    []byte
	wbuf    []byte
	bytebuf *bytes.Buffer
}

func NewConn(c io.ReadWriteCloser) *Conn {
	m := &Conn{}
	m.rwc = c
	m.rw = c.(io.ReadWriter)
	m.bytebuf = bytes.NewBuffer(m.wbuf)
	return m
}

func (m *Conn) ReadMsg() (id byte, header, payload []byte, err error) {
	id, err = m.readMessageType()
	if err != nil {
		return
	}
	info, ok := infoMap[id]
	if !ok {
		err = fmt.Errorf("conn: unknown command with id=%v", id)
		return
	}
	header, err = m.readMessageHeader(info.HeaderLen)
	if err != nil {
		return
	}
	payload, err = m.readMessagePayload(info.LenCalc(header))
	return
}

func (m *Conn) readMessageType() (byte, error) {
	_, err := io.ReadAtLeast(m.rw, m.header[:1], 1)
	return m.header[0], err
}

func (m *Conn) readMessageHeader(size int) ([]byte, error) {
	n, err := io.ReadAtLeast(m.rw, m.header[:size], size)
	return m.header[:n], err
}

func (m *Conn) readMessagePayload(size int) ([]byte, error) {
	if cap(m.rbuf) < size {
		m.rbuf = make([]byte, size)
	}
	n, err := io.ReadAtLeast(m.rw, m.rbuf[:size], size)
	return m.rbuf[:n], err
}

func (m *Conn) WriteValues(a ...interface{}) error {
	for _, i := range a {
		err := binary.Write(m.bytebuf, binary.BigEndian, i)
		if err != nil {
			return (err)
		}
	}
	return nil
}

func (m *Conn) RegisterCommand(id byte, headerlen int,
	lencalc func([]byte) int, name string) {
	if infoMap == nil {
		infoMap = make(map[byte]messageInfo)
	}
	infoMap[id] = messageInfo{name, headerlen, lencalc}
}

func (m *Conn) Name(id byte) string {
	return infoMap[id].Name
}

func (m *Conn) Reset() {
	m.bytebuf.Reset()
}

func (m *Conn) Flush() error {
	_, err := m.rw.Write(m.bytebuf.Bytes())
	return err
}
