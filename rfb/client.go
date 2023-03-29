package rfb

import (
	"image"
	"io"
	"log"
	"rfbserver/rfb/encoder"
	"sync"
	"time"

	//	"rfbserver/desktop/input"
	"rfbserver/region"
)


var id int

type Client struct {
	connLock	sync.Mutex
	conn *Conn
	s               *Server
	encoder         encoder.RfbEncoder
	losslessEncoder encoder.RfbEncoder
	id              int
	lastTime        time.Time
	prevFb          *LockableImage
	listener        chan bool
	dirt            region.Region
	lossless        region.Region
}

func NewClient(s *Server, conn io.ReadWriteCloser, bounds image.Rectangle, listener chan bool) *Client {
	c := Client{
		sync.Mutex{},
		NewConn(conn),
		s,
		encoder.NewTightEncoder(10),
		encoder.NewTightEncoder(100),
		id,
		time.Now(),
		&LockableImage{Image: image.NewRGBA(bounds)},
		listener,
		region.NewRegion(),
		region.NewRegion(),
	}
	id++
	return &c;
}

// Serve a new connection.
func (cm *Client) serve() {
	var err error
	defer cm.Close()
	defer cm.s.RemoveListener(cm.listener)

	log.Printf("client %d connected", cm.id)

	if !cm.handshake() {
		log.Printf("client %d can't handshake", cm.id)
		return
	}

	log.Printf("client %d handshaked successfully", cm.id)

	for {
		id, _, payload, err := cm.conn.ReadMsg()
		if err != nil {
			break
		}
		h, ok := messageHandlers[id]
		if !ok {
			log.Printf("client %d: handler not registered for command with id=%v", cm.id, id)
			continue
		}
		h.Handler(payload, cm)
	}
	log.Printf("client %d closing connection: %v\n", cm.id, err)

}

func (cm *Client) handshake() bool {
	//ProtocolVersion:
	v, ok := serverProtocolVersionHandler(cm.conn.rw)
	if ok != nil {
		log.Printf("client %d error in Protocol version: %v", cm.id, ok)
		//		return false
	}
	if v != v8 {
		log.Printf("client %d protocol version not supported: %s", cm.id, v)
		//		return false
	}

	log.Printf("client %d protocol version: %s", cm.id, v)

	//SecurityType:
	st, ok := serverSecurityTypeHandler(cm.conn.rw)
	if ok != nil {
		log.Printf("client %d error reading SecurityType", cm.id)
		return false
	}
	log.Printf("client %d security types: %v", cm.id, st)
	//SecurityResult:
	serverSecurityResultHandler(cm.conn.rw)
	//Init:
	isShared, _ := serverInitHandler(cm.conn.rw, cm.s.Config)
	log.Printf("client %d is shared: %v", cm.id, isShared)
	return true
}

func (cm *Client) PushFrame(requestRect image.Rectangle, incremental bool) {
	last := cm.s.lastFb
	prev := cm.prevFb

	if incremental {
		<-cm.listener
	} else {
		prev.Image = &image.RGBA{}
		cm.lossless = region.NewRegion()
	}
	imp := prev.Image.SubImage(requestRect)
	last.RLock()
	iml := last.Image.SubImage(requestRect)
	select {
	case cm.s.updateRequest <- true:
	default:
	}

	blocks := SliceScreen10(iml, imp)
	prev.Image = last.Image

	if len(blocks) != 0 {
		newDirt := region.NewRegionFromRects(blocks)
		cm.dirt.Add(newDirt)
		cm.lossless.Subtract(newDirt)
	}
	b := last.Image.Bounds()
	screenArea := b.Dx() * b.Dy()
	lossless := TakeSomeRectsFromRegion(&cm.lossless, screenArea/100)
	if len(lossless) == 0 {
		cm.lossless.Set(cm.dirt)
		cm.dirt = region.NewRegion()
	}

	// TODO change to pipeline?
	dataLoss := make(chan encoder.PackedRect)
	dataLossless := make(chan encoder.PackedRect)
	for _, r := range blocks {
		im := last.Image.SubImage(r)
		go func() {
			dataLoss <- cm.encoder.Encode(im)
		}()
	}
	for _, r := range lossless {
		im := last.Image.SubImage(r)
		go func() {
			dataLossless <- cm.losslessEncoder.Encode(im)
		}()
	}
	var rects []encoder.PackedRect
	var rectsLossless []encoder.PackedRect
	for i := 0; i < len(blocks); i++ {
		rects = append(rects, <-dataLoss)
	}
	for i := 0; i < len(lossless); i++ {
		rectsLossless = append(rectsLossless, <-dataLossless)
	}
	last.RUnlock()
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	cm.conn.Reset()
	// Send header
	cm.conn.WriteValues(uint8(CmdFramebufferUpdate),
		uint8(0),
		uint16(len(rects)+len(rectsLossless)),
	)
	for _, r := range rects {
		// Send rectangle:
		cm.conn.WriteValues(uint16(r.Rect.Min.X), uint16(r.Rect.Min.Y),
			uint16(r.Rect.Dx()), uint16(r.Rect.Dy()),
			int32(r.Encoding),
		)
		cm.conn.WriteValues(r.Data)
	}
	sizeLoss := cm.conn.bytebuf.Len()
	for _, r := range rectsLossless {
		// Send rectangle:
		cm.conn.WriteValues(uint16(r.Rect.Min.X), uint16(r.Rect.Min.Y),
			uint16(r.Rect.Dx()), uint16(r.Rect.Dy()),
			int32(r.Encoding),
		)
		cm.conn.WriteValues(r.Data)
	}
	cm.conn.Flush()
	sizeLossless := cm.conn.bytebuf.Len() - sizeLoss
	cm.s.stat <- Stat{time.Since(cm.lastTime), sizeLoss, sizeLossless}
	cm.lastTime = time.Now()
}

func (cm *Client) Close() {
	cm.conn.rwc.Close()
}