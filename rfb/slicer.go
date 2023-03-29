package rfb

import (
	"bytes"
	"image"
//	"log"
	"math/bits"
	"rfbserver/region"
)

type Bitmask struct {
	data []uint64
	bounds image.Rectangle
	dx int
	dy int
}

func SliceScreen(s1, s2 *image.RGBA) []image.Rectangle {
	b := s1.Bounds()
	if s1.Rect != s2.Rect {
		return []image.Rectangle{b}
	}
	type pair struct {
		start, end int
	}
	var stripes []pair
	var on bool 
	var start int
	for y := b.Min.Y; y < b.Max.Y; y++ {
		from := s1.PixOffset(b.Min.X, y)
		to := s1.PixOffset(b.Max.X, y)
		l1 := s1.Pix[from:to]
		l2 := s2.Pix[from:to]
		if !bytes.Equal(l1, l2) {
			if !on {
				on = true
				start = y
				y += 16 // why not?
			}
		} else {
			if on {
				stripes = append(stripes, pair{start, y})
				on = false
			}
		}
	}
	if on {
		stripes = append(stripes, pair{start, b.Max.Y})
		on = false
	}

	if len(stripes) == 0 {
		return nil
	}

	rr := make([]image.Rectangle, len(stripes))
	rr = rr[:0]
	for _, s := range stripes {
		r := image.Rect(b.Min.X,s.start, b.Max.X, s.end)
		if r.Empty() {
			continue
		}
		rr = append(rr, r)		
	}
	return rr
}

func SliceScreen10(s1, s2 image.Image) []image.Rectangle {
	const N = 10
	rgba1 := *s1.(*image.RGBA)
	rgba2 := *s2.(*image.RGBA)
	b := rgba1.Bounds()
	dX := b.Dx() / N

	var rr []image.Rectangle
	
	data := make(chan []image.Rectangle)

	for i := 0; i < N ; i++ {
		X := i * dX
		r := image.Rect(X, b.Min.Y, X+dX, b.Max.Y)
		si1 := rgba1.SubImage(r).(*image.RGBA)
		si2 := rgba2.SubImage(r).(*image.RGBA)
		go func () {
			data <- SliceScreen(si1, si2)
		}()
	}	
	for i := 0; i < N; i++ {
		rr = append(rr, <-data...)
	}
	return rr
}	

var mask1 = [65]uint64{
	0x0000000000000000,
	0x0000000000000001,
	0x0000000000000003,
	0x0000000000000007,
	0x000000000000000f,
	0x000000000000001f,
	0x000000000000003f,
	0x000000000000007f,
	0x00000000000000ff,
	0x00000000000001ff,
	0x00000000000003ff,
	0x00000000000007ff,
	0x0000000000000fff,
	0x0000000000001fff,
	0x0000000000003fff,
	0x0000000000007fff,
	0x000000000000ffff,
	0x000000000001ffff,
	0x000000000003ffff,
	0x000000000007ffff,
	0x00000000000fffff,
	0x00000000001fffff,
	0x00000000003fffff,
	0x00000000007fffff,
	0x0000000000ffffff,
	0x0000000001ffffff,
	0x0000000003ffffff,
	0x0000000007ffffff,
	0x000000000fffffff,
	0x000000001fffffff,
	0x000000003fffffff,
	0x000000007fffffff,
	0x00000000ffffffff,
	0x00000001ffffffff,
	0x00000003ffffffff,
	0x00000007ffffffff,
	0x0000000fffffffff,
	0x0000001fffffffff,
	0x0000003fffffffff,
	0x0000007fffffffff,
	0x000000ffffffffff,
	0x000001ffffffffff,
	0x000003ffffffffff,
	0x000007ffffffffff,
	0x00000fffffffffff,
	0x00001fffffffffff,
	0x00003fffffffffff,
	0x00007fffffffffff,
	0x0000ffffffffffff,
	0x0001ffffffffffff,
	0x0003ffffffffffff,
	0x0007ffffffffffff,
	0x000fffffffffffff,
	0x001fffffffffffff,
	0x003fffffffffffff,
	0x007fffffffffffff,
	0x00ffffffffffffff,
	0x01ffffffffffffff,
	0x03ffffffffffffff,
	0x07ffffffffffffff,
	0x0fffffffffffffff,
	0x1fffffffffffffff,
	0x3fffffffffffffff,
	0x7fffffffffffffff,
	0xffffffffffffffff,
}

func setBits(start int, end int) uint64 {
	out := mask1[end] &^ mask1[start]
//	log.Printf("start: %d, end: %d mask: %b", start, end, out)
	return out
}

func NewBitmask(bounds image.Rectangle)  Bitmask {
	dx := bounds.Dx() / 64
	if dx < 16 {
		dx = 16
	}
	ny := bounds.Dy() / dx
	dy := bounds.Dy() / ny
	if dy < 16 {
		dy = 16
	}
	mask := make([]uint64, bounds.Dy() / dy)
	return Bitmask{mask, bounds, dx, dy}
}	


func RectsToBitmask(bounds image.Rectangle, dirtyRects []image.Rectangle)  Bitmask {
	out := NewBitmask(bounds)
	for _, r := range dirtyRects {
		start := r.Min.Y / out.dy
		out.data[start] |= setBits(r.Min.X / out.dx, (r.Max.X - 1) / out.dx)
		for i := start + 1 ; i < r.Max.Y / out.dy; i++ {
			out.data[i] |= setBits(r.Min.X / out.dx, (r.Max.X - 1) / out.dx)
		}
	}
	return out
}	

// this function changes mask
func BitmaskRowToRect(startRow int, mask Bitmask, maxsquires int)  (r image.Rectangle, processed int) {
	dx := mask.dx
	dy := mask.dy
	endRow := len(mask.data)
	word := mask.data[startRow]
	if word == 0 {
		return r, 0
	}
	start := bits.TrailingZeros64(word)
	end := 64 - bits.LeadingZeros64(word)
	for k := start + 1; k < end; k++ {
		if word & (1 << uint(k)) == 0 {
			end = k
			break
		}
	}
	r.Min.Y = startRow * dy
	r.Max.Y = r.Min.Y + dy
	r.Min.X = start * dx
	r.Max.X = end * dx
	bitRow := setBits(start, end)
	mask.data[startRow] &^= bitRow
	processed = end - start
	for j := startRow + 1; j < endRow; j++ {
		if processed > maxsquires {
			break
		}
		if (mask.data[j] & bitRow) != bitRow {
			break
		}
		r.Max.Y += dy
		mask.data[j] &^= bitRow
		processed += end - start
	}
	return r, processed
}	

func TakeSomeRectsFromBitmask(mask Bitmask, part int)  []image.Rectangle {
	maxsquares := len(mask.data) * len(mask.data) / part
	if maxsquares < 1 {
		maxsquares = 1
	}
	rects := []image.Rectangle{}
	
	for row, data := range mask.data {
		if data == 0 {
			continue
		}
		for ; maxsquares > 0; {
			r, processed := BitmaskRowToRect(row, mask, maxsquares)
			if processed == 0 {
				break
			}
			maxsquares -= processed
			rects = append(rects, r)
		}
	}
	return rects
}

func BitmaskOr(mask Bitmask, tool Bitmask)  Bitmask {
	for i, word := range tool.data {
		mask.data[i] |= word
	}
	return mask
}	

func BitmaskAndNot(mask Bitmask, tool Bitmask)  Bitmask {
	for i, word := range tool.data {
		mask.data[i] &^= word
	}
	return mask
}	

func CountOnes(mask Bitmask) int {
	out := 0
	for _, word := range mask.data {
		out += bits.OnesCount64(word)
	}
	return out
}	

func TakeSomeRectsFromRegion(reg *region.Region, area int)  (out []image.Rectangle) {
	rects := reg.GetRectVector()
	for i := len(rects) - 1 ; i >= 0 && area > 0 ; i-- {
		r := rects[i]
		a := r.Dx() * r.Dy()
		if a > area {
			r = image.Rect(r.Min.X, r.Min.Y, r.Min.X + r.Dx() * area / a, r.Max.Y)
			a = r.Dx() * r.Dy()
		} 
		area -= a
		reg.Subtract(region.NewRegionFromRect(r))
		out = append(out, r)
	} 
	return
}