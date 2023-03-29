package encoder

import (
//	"time"
	"log"
	"image"
)


type HextileEncoder Encoder

func NewHextileEncoder() *HextileEncoder {
	return &HextileEncoder{}
}

func (this HextileEncoder) Id() int {
	return encodingHextile
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func (this HextileEncoder) Encode(im image.Image) []PackedRect {
	log.Printf("hextile: encoding")
  	var t image.Rectangle
  
	rects := []PackedRect{}
	var output [256*4 + 1]uint8
	var oldBg, oldFg  PIXEL_T 
  	var oldBgValid, oldFgValid bool 

	tile := NewHextileTile()
	source := im.(*image.RGBA)
  	r := source.Bounds()
	
  	encoded := make([]uint8, 256 * sizeof_PIXEL_T)
	

  	for t.Min.Y = r.Min.Y; t.Min.Y < r.Max.Y; t.Min.Y += 16 {
//	log.Printf("hextile: Y: %d", t.Min.Y)
//    	t.Max.Y = min(r.Max.Y, t.Min.Y + 16)
    	for t.Min.X = r.Min.X; t.Min.X < r.Max.X; t.Min.X += 16 {
//      		t.Max.X = min(r.Max.X, t.Min.X + 16)
    		tile.NewTile(*source, t)
      		tileType := tile.GetFlags()
      		encodedLen := tile.GetSize()
//  log.Print("tileType:",tileType, " encodedLen: ", encodedLen )
      		if tileType & HEXTILE_RAW != 0 ||
           	   encodedLen >= t.Dx() * t.Dy() * sizeof_PIXEL_T {
      			output[0] = HEXTILE_RAW
				for i := 0; i < t.Dx() * t.Dy(); i++ {
					j := i*4 + 1
					b := uint32(tile.tile[i])
	      			output[j]   = uint8(0xff & (b >> 24))
	      			output[j+1] = uint8(0xff & (b >> 16))
	      			output[j+2] = uint8(0xff & (b >> 8))
	      			output[j+3] = uint8(0xff & b)
				}
       			oldBgValid = false
				oldFgValid = false
//  log.Print("output:", output)
				rects = append(rects, PackedRect{	Rect: t, 
													Encoding: encodingHextile, 
													Data: output[:]})
       			continue
        	}

      		bg := tile.GetBackground()
      		fg := PIXEL_T(0)

      		if !oldBgValid || oldBg != bg {
        		tileType |= HEXTILE_BG_SPECIFIED
        		oldBg = bg
        		oldBgValid = true
      		}

      		if (tileType & HEXTILE_ANY_SUBRECTS) != 0 {
        		if (tileType & HEXTILE_SUBRECTS_COLOURED) != 0 {
          			oldFgValid = false;
        		} else {
          			fg = tile.GetForeground()
          			if (!oldFgValid || oldFg != fg) {
            			tileType |= HEXTILE_FG_SPECIFIED
            			oldFg = fg
            			oldFgValid = true
          			}
        		}
				
        		tile.Encode(encoded)
      		}
			i := 0
      		output[i] = tileType; i++
      		if (tileType & HEXTILE_BG_SPECIFIED) != 0 {
      			output[i] = uint8(0xff & (uint32(bg)>>24)); i++
      			output[i] = uint8(0xff & (uint32(bg)>>16)); i++
      			output[i] = uint8(0xff & (uint32(bg)>>8)); i++
      			output[i] = uint8(0xff & uint32(bg)); i++
      		}
      		if (tileType & HEXTILE_FG_SPECIFIED) != 0 {
      			output[i] = uint8(0xff & (uint32(fg)>>24)); i++
      			output[i] = uint8(0xff & (uint32(fg)>>16)); i++
      			output[i] = uint8(0xff & (uint32(fg)>>8)); i++
      			output[i] = uint8(0xff & uint32(fg)); i++
      		}
			slice := output[:i]
		    if (tileType & HEXTILE_ANY_SUBRECTS) != 0 {
				slice = append(slice, encoded[:encodedLen]...)
      		}
			
			rects = append(rects, PackedRect{Rect: t, Data: slice})
    	}
  	}
	log.Printf("hextile: rects number %d", len(rects))
	return rects
}

func append4byte(a []uint8, w uint32) []uint8 {
	return append(a, uint8(0xff&(w>>24)), uint8(0xff&(w>>16)),
				  	 uint8(0xff&(w>>8)),  uint8(0xff&w))
}

func (this *HextileEncoder) SetPF(pf PixelFormat) {
	this.pf = pf
}
