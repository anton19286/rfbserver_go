package encoder

import (
	"image"
)
const (
	// framebuffer encodings
	encodingRaw      = 0
	encodingCopyRect = 1
	encodingHextile  = 5
	encodingTight    = 7
	encodingTightPng = -260

)
type PixelFormat struct {
	BPP, Depth                      uint8
	BigEndian, TrueColour           uint8
	RedMax, GreenMax, BlueMax       uint16
	RedShift, GreenShift, BlueShift uint8
}

type Encoder struct {
	pf PixelFormat
}

type PackedRect struct {
	Rect image.Rectangle
	Encoding int32
	Data []byte
}

type RfbEncoder interface {
	Encode(image.Image) PackedRect
	Id() int
	SetPF(PixelFormat)
}

type RfbDecoder interface {
	Decode([]PackedRect) image.Image
}

var DefaultPixelFormat = PixelFormat{
	BPP:        32,
	Depth:      24,
	BigEndian:  0,
	TrueColour: 1,
	RedMax:     0xff,
	GreenMax:   0xff,
	BlueMax:    0xff,
	RedShift:   16,
	GreenShift: 8,
	BlueShift:  0,
}
