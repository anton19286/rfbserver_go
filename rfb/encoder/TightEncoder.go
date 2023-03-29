package encoder

import (
	"bytes"
	"image"
	"image/jpeg"
)


type TightEncoder struct {
	Encoder
	Quality int 	
}	

func NewTightEncoder(quality int) *TightEncoder {
	return &TightEncoder{Quality : quality}
}

func (this TightEncoder) Id() int {
	return encodingTight
}

func (this TightEncoder) Encode(im image.Image) PackedRect {
	return this.jpegEncode(im.(*image.RGBA))
}

func (this TightEncoder) jpegEncode(im *image.RGBA) PackedRect{
	buf := new(bytes.Buffer)
	var o jpeg.Options
	o.Quality = this.Quality
	ok := jpeg.Encode(buf, im, &o)
	if ok != nil {
		return PackedRect{Rect: image.Rectangle{}, Encoding: encodingTight, Data: nil}
	}
	jpegLen := buf.Len()
	out1 := make([]byte, jpegLen + 4)
	out := out1[:0]
	
	out = append(out, 0x90) //Tight jpeg
	if jpegLen < 128 {
		out = append(out, uint8(0x7f & jpegLen))
	} else if jpegLen < 16383 {
		out = append(out, uint8(0x80 + (0x7f & jpegLen)))
		out = append(out, uint8(0x7f & (jpegLen >> 7)))
	} else {
		out = append(out, uint8(0x80 + (0x7f & jpegLen)))
		out = append(out, uint8(0x80 + (0x7f & (jpegLen >> 7))))
		out = append(out, uint8(0xff & (jpegLen >> 14)))
	}
	out = append(out, buf.Bytes()...)
	return PackedRect{Rect: im.Bounds(), Encoding: encodingTight, Data: out}
}

func (this *TightEncoder) SetPF(pf PixelFormat) {
	this.pf = pf
}
