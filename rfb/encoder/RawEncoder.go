package encoder

import (
	"image"
	"log"
)

type RawEncoder Encoder

func NewRawEncoder() *RawEncoder {
	e := RawEncoder{}
	e.pf = DefaultPixelFormat
	return &e
}


func (this RawEncoder) Id() int {
	return encodingRaw
}

func (this RawEncoder) Encode(im image.Image) []PackedRect {
	var buf []byte
	pf := this.pf
	if pf.BPP == 32 {
		buf = this.encodeRGBA32bpp(im.(*image.RGBA), pf)
	} else {
		buf = this.encodeGeneric(im, pf)
	}
	return []PackedRect{PackedRect{Rect: im.Bounds(), Encoding: encodingRaw, Data: buf}}
}

func (this RawEncoder) encodeRGBA32bpp(im *image.RGBA, pf PixelFormat) []byte {
	var r, g, b byte
	buf := make([]byte, len(im.Pix))
	out := buf
	if pf.BigEndian != 0 {
		for i, v8 := range im.Pix {
			switch i % 4 {
				case 0: // red
					r = v8 //redshift of 10
				case 1: // green
					g = v8 // redshift of 8
				case 2: // blue
					b = v8
				case 3: // alpha, unused.  use this to just move the dest
					out[0] = 0
					out[1] = r
					out[2] = g
					out[3] = b
					out = out[4:]
			}
		}
	} else {
		for i, v8 := range im.Pix {
			switch i % 4 {
				case 0: // red
					r = v8 //redshift of 10
				case 1: // green
					g = v8 // redshift of 8
				case 2: // blue
					b = v8
				case 3: // alpha, unused.  use this to just move the dest
					out[0] = b
					out[1] = g
					out[2] = r
					out[3] = 0
					out = out[4:]
			}
		}
	}
	log.Printf("sending raw %d bytes of %v image", len(im.Pix), im.Bounds())
	return buf
}

func (this RawEncoder) encodeGeneric(im image.Image, pf PixelFormat) []byte { 
	log.Panic("Generic RawEncoder does not implemented, requested PF:", pf)
	return []byte{} 
}

func (this RawEncoder) SetPF(pf PixelFormat) {
	this.pf = pf
}
