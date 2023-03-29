package encoder

import (
	"log"
	"image"
)


const HEXTILE_RAW = 1
const HEXTILE_BG_SPECIFIED = 2
const HEXTILE_FG_SPECIFIED = 4
const HEXTILE_ANY_SUBRECTS = 8
const HEXTILE_SUBRECTS_COLOURED = 16

type HextileTile struct {
	tile [256]PIXEL_T
  	width int
  	height int 

  	size int
  	flags uint8
  	background PIXEL_T
  	foreground PIXEL_T

  	numSubrects int
  	coords [256 * 2]uint8
  	colors [256]PIXEL_T 
	pal TightPalette 
}	

func NewHextileTile() *HextileTile {
	h := &HextileTile{}
	h.pal = NewTightPalette(48+2*sizeof_PIXEL_T)
	return h
}

// Initialize existing object instance with new tile data.
func (h *HextileTile) NewTile(src image.RGBA, pos image.Rectangle)  {
	for j := 0; j < pos.Dy() ; j++ {
		for i := 0; i < pos.Dx() ; i++ {
			o := pos.Min.X + i * 4  + (pos.Min.Y + j) * src.Stride
			h.tile[j * 16 + i] = PIXEL_T(	uint32(src.Pix[o])<<24 + 
											uint32(src.Pix[o+1])<<16 + 
											uint32(src.Pix[o+2])<<8 + 
											uint32(src.Pix[o+3]))
		}
	}	
	h.width = pos.Dx()
	h.height = pos.Dy()
	h.Analize()
}

// Flags can include: HEXTILE_RAW, HEXTILE_ANY_SUBRECTS and
// HEXTILE_SUBRECTS_COLOURED. Note that if HEXTILE_RAW is set, other
// flags make no sense. Also, HEXTILE_SUBRECTS_COLOURED is meaningful
// only when HEXTILE_ANY_SUBRECTS is set as well.
func (h HextileTile) GetFlags() uint8 { 
	return h.flags
}

// Returns the size of encoded subrects data, including subrect count.
// The size is zero if flags do not include HEXTILE_ANY_SUBRECTS.
func (h HextileTile) GetSize() int {
	return h.size
}

// Return optimal background.
func (h HextileTile) GetBackground() PIXEL_T { 
	return h.background 
}

// Return foreground if flags include HEXTILE_SUBRECTS_COLOURED.
func (h HextileTile) GetForeground() PIXEL_T { 
	return h.foreground
}

func (h *HextileTile) Analize() {
	var	processed [16][16]bool
  	if h.width == 0 || h.height == 0 {
		log.Panic("hextile: Analize")
	}
// log.Print("in:",*h)
	color := h.tile[0]
	i := 1
	l := h.width * h.height
	for i < l && h.tile[i] == color {
		i++ 
	}
 	// Handle solid tile	
	if i == l {
		h.background = color
		h.flags = 0
		h.size = 0	
		return
	}
	// Compute number of complete rows of the same color, at the top
	y := i / h.width	  	
	h.pal.Reset()
  	h.numSubrects = 0

  	// Have we found the first subrect already?
	var colorsPtr, coordsPtr int
  	if y > 0 {
    	h.colors[colorsPtr] = color; colorsPtr++ 
    	h.coords[coordsPtr] = 0; coordsPtr++
    	h.coords[coordsPtr] = uint8(((h.width - 1) << 4) | ((y - 1) & 0x0F))
		coordsPtr++
    	h.pal.Insert(color, 1)
	    h.numSubrects++
  	}

  	var x, sx, sy, sw, sh, max_x int

  	for ; y < h.height; y++ {
    	for x = 0; x < h.width; x++ {
      		// Skip pixels that were processed earlier
      		if (processed[y][x]) {
        		continue
      		}
      		// Determine dimensions of the horizontal subrect
		    color = h.tile[y * h.width + x]
      		for sx = x + 1; sx < h.width; sx++ {
      			if h.tile[y * h.width + sx] != color {
	        		break
				}
      		}
      		sw = sx - x
      		max_x = sx
      		for sy = y + 1; sy < h.height; sy++ {
        		for sx = x; sx < max_x; sx++ {
          			if h.tile[sy * h.width + sx] != color {
            			goto done
					}
        		}
      		}
done:
      		sh = sy - y;

      		// Save properties of this subrect
     		h.colors[colorsPtr] = color; colorsPtr++
      		h.coords[coordsPtr] = uint8((x << 4) | (y & 0x0F)); coordsPtr++
      		h.coords[coordsPtr] = uint8(((sw - 1) << 4) | ((sh - 1) & 0x0F))
			coordsPtr++

      		if (h.pal.Insert(color, 1) == 0) {
        		// Handle palette overflow
        		h.flags = HEXTILE_RAW
        		h.size = 0
// log.Print("out:",*h)
        		return
      		}

      		h.numSubrects++

      		// Mark pixels of this subrect as processed, below this row
      		for sy = y + 1; sy < y + sh; sy++ {
        		for sx = x; sx < x + sw; sx++ {
          			processed[sy][sx] = true
				}
      		}

      		// Skip processed pixels of this row
      		x += sw - 1
    	}
  	}

  	// Save number of colors in this tile (should be no less than 2)
  	numColors := h.pal.GetNumColors()
	if !(numColors >= 2) {
		log.Print(h.pal)
	    log.Panic("hextile: numColors < 2")
	}

  	h.background = h.pal.GetEntry(0)
  	h.flags = HEXTILE_ANY_SUBRECTS
  	numSubrects := h.numSubrects - h.pal.GetCount(0)

  if numColors == 2 {
    // Monochrome tile
    h.foreground = h.pal.GetEntry(1);
    h.size = 1 + 2 * numSubrects
  } else {
    // Colored tile
    h.flags |= HEXTILE_SUBRECTS_COLOURED
    h.size = 1 + (2 + sizeof_PIXEL_T) * numSubrects
  }
//log.Print("out:",*h)

}

func (h *HextileTile) Encode(dst []uint8) {
	if h.numSubrects == 0 || h.flags & HEXTILE_ANY_SUBRECTS == 0 {
	    log.Panic("hextile: Encode")
	}

  	// Zero subrects counter
  	dst[0] = 0

	d := 1

  	for i := 0; i < h.numSubrects; i++ {
    	if h.colors[i] == h.background {
      		continue;
    	}
    	if h.flags & HEXTILE_SUBRECTS_COLOURED != 0 {
      		for j := 0; j < sizeof_PIXEL_T; j++ {
				dst[d] = uint8(0xff & (uint32(h.colors[i]) >> uint(8 * j)))
				d++
			}
    	}
    	dst[d] = h.coords[i * 2]; d++
    	dst[d] = h.coords[i * 2 + 1]; d++

    	dst[0]++
  	}

  	if d != h.size {
		log.Panic("hextile: Encode")
	}
}
