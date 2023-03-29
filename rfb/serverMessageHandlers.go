package rfb

import (
	"encoding/binary"
	"image"
	"log"
	"rfbserver/rfb/encoder"
)

var messageHandlers map[byte]messageHandler

func init() {
	messageHandlers = make(map[byte]messageHandler)
	messageHandlers[CmdSetPixelFormat] = messageHandler{
		Id:      CmdSetPixelFormat,
		Handler: cmdSetPixelFormatHandler,
		Name:    "CmdSetPixelFormat",
	}

	messageHandlers[CmdSetEncodings] = messageHandler{
		Id:      CmdSetEncodings,
		Handler: cmdSetEncodingsHandler,
		Name:    "CmdSetEncodings",
	}

	messageHandlers[CmdFramebufferUpdateRequest] = messageHandler{
		Id:      CmdFramebufferUpdateRequest,
		Handler: cmdFramebufferUpdateRequestHandler,
		Name:    "CmdFramebufferUpdateRequest",
	}

	messageHandlers[CmdKeyEvent] = messageHandler{
		Id:      CmdKeyEvent,
		Handler: cmdKeyEventHandler,
		Name:    "CmdKeyEvent",
	}

	messageHandlers[CmdPointerEvent] = messageHandler{
		Id:      CmdPointerEvent,
		Handler: cmdPointerEventHandler,
		Name:    "CmdPointerEvent",
	}

	messageHandlers[CmdClientCutText] = messageHandler{
		Id:      CmdClientCutText,
		Handler: cmdClientCutTextHandler,
		Name:    "CmdClientCutText",
	}
}

func cmdSetPixelFormatHandler(pl []byte, cm *Client) {
	if len(pl) != 16 {
		log.Printf("rfb: SetPixelFormatHandler: wrong data length %d", len(pl))
		return
	}
	pf := encoder.PixelFormat{
		BPP:        pl[0],
		Depth:      pl[1],
		BigEndian:  pl[2],
		TrueColour: pl[3],
		RedMax:     uint16(pl[4]<<8) + uint16(pl[5]),
		GreenMax:   uint16(pl[6]<<8) + uint16(pl[7]),
		BlueMax:    uint16(pl[8]<<8) + uint16(pl[9]),
		RedShift:   pl[10],
		GreenShift: pl[11],
		BlueShift:  pl[12],
	}
	//	cm.Config.PixelFormat = pf
	cm.encoder.SetPF(pf)
	cm.losslessEncoder.SetPF(pf)
	log.Printf("rfb: client asks pixel format:  %v\n", pf)
}

func cmdSetEncodingsHandler(pl []byte, cm *Client) {
	if len(pl)%4 != 0 {
		log.Printf("rfb: SetEncodingsHandler: wrong data length %d", len(pl)%4)
		return
	}
	e := []int32{}
	for i := 0; i < len(pl); i += 4 {
		e = append(e, int32(rU32(pl[i:])))
	}
	//	cm.Config.clientEncodings = e
	log.Printf("rfb: client asks encodings:  %v\n", e)
}

var rU16 = binary.BigEndian.Uint16

func rU32(b []byte) uint32 {
	return uint32(b[0]<<24) + uint32(b[1]<<16) + uint32(b[2]<<8) + uint32(b[3])
}

func cmdFramebufferUpdateRequestHandler(pl []byte, cm *Client) {
	incremental := pl[0] != 0
	x := int(rU16(pl[1:]))
	y := int(rU16(pl[3:]))
	width := int(rU16(pl[5:]))
	height := int(rU16(pl[7:]))
	r := image.Rect(x, y, x+width, y+height)
	//	log.Printf("rfb: FramebufferUpdateRequest: rect:%v incremental: %v\n",
	//				r, incremental)
	go cm.PushFrame(r, incremental)
}

func cmdKeyEventHandler(pl []byte, cm *Client) {
	log.Printf("rfb: KeyEventHandler: %v\n", pl)
}

func cmdPointerEventHandler(pl []byte, cm *Client) {
	//	buttonMask := pl[0]
	//	x := int(rU16(pl[1:]))
	//	y := int(rU16(pl[3:]))
	//	cm.s.userInput.SetMouseEvent(image.Point{x, y}, buttonMask)
	//
	// log.Printf("rfb: PointerEventHandler: %v\n", pl)
}
func cmdClientCutTextHandler(pl []byte, cm *Client) {
	log.Printf("rfb: ClientCutTextHandler: %v\n", pl)
}

func (s *Server) RegisterHandler(h messageHandler) {
	if messageHandlers == nil {
		messageHandlers = make(map[byte]messageHandler)
	}
	messageHandlers[h.Id] = h
}
