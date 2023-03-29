package encoder

import (
	"image/png"
	"bytes"
	"image"
//	"log"
)


type TightPngEncoder Encoder

func NewTightPngEncoder() *TightPngEncoder {
	return &TightPngEncoder{}
}

func (this TightPngEncoder) Id() int {
	return encodingTightPng
}

func (this TightPngEncoder) Encode(im image.Image) PackedRect {
	return this.pngEncode(im.(*image.RGBA))
}

func (this TightPngEncoder) pngEncode(im *image.RGBA) PackedRect {
	buf := new(bytes.Buffer)
	ok := png.Encode(buf, im)
	if ok != nil {
		return PackedRect{Rect: image.Rectangle{}, Encoding: encodingTight, Data: nil}
	}
	pngLen := buf.Len()
	out1 := make([]byte, pngLen + 4)
	out := out1[:0]
	
	out = append(out, 0xa0) //Tight png
	if pngLen < 128 {
		out = append(out, uint8(0x7f & pngLen))
	} else if pngLen < 16383 {
		out = append(out, uint8(0x80 + (0x7f & pngLen)))
		out = append(out, uint8(0x7f & (pngLen >> 7)))
	} else {
		out = append(out, uint8(0x80 + (0x7f & pngLen)))
		out = append(out, uint8(0x80 + (0x7f & (pngLen >> 7))))
		out = append(out, uint8(0xff & (pngLen >> 14)))
	}
	out = append(out, buf.Bytes()...)
//	log.Printf("sending png %d bytes of %v image", len(out1), im.Bounds())

	return PackedRect{Rect: im.Bounds(), Encoding: encodingTightPng, Data: out}
	
}

func (this *TightPngEncoder) SetPF(pf PixelFormat) {
	this.pf = pf
}
