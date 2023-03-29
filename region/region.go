package region
// #cgo CFLAGS: -D DONT_INLINE_REGION_OPS
// #include "x11region.h"
import 	"C"

import (
	"image"
	"unsafe"
)


type Region struct {
	reg C.struct__Region
} 

func NewRegion() (reg Region) {
  	C.miRegionInit(&reg.reg, C.NullBox, 0)
	return reg
}

func NewRegionFromRect(rect image.Rectangle) (reg Region) {
  	if !rect.Empty() {
  	  	var box C.struct__Box
  	  	box.x1 = C.short(rect.Min.X)
  	  	box.x2 = C.short(rect.Max.X)
  	  	box.y1 = C.short(rect.Min.Y)
  	  	box.y2 = C.short(rect.Max.Y)
  	  	C.miRegionInit(&reg.reg, &box, 0)
  	} else {
	    C.miRegionInit(&reg.reg, C.NullBox, 0)
	}
	return reg
}

func NewRegionFromRects(rects []image.Rectangle) Region {
	reg := NewRegion()
	for _, r := range rects {
		reg.Add(NewRegionFromRect(r))
	}
	return reg
}

func CopyRegion(src Region) Region {
	reg := NewRegion()
	reg.Set(src)
	return reg
}

func (r * Region) Close() {
  	C.miRegionUninit(&r.reg)
}


func (r * Region) Clear() {
  	C.miRegionEmpty(&r.reg)
}

func (r * Region) Set(src Region) {
  	C.miRegionCopy(&r.reg, &src.reg)
}


func (r * Region) AddRect(rect image.Rectangle) {
  	if !rect.Empty() {
		temp := NewRegionFromRect(rect)
    	r.Add(temp)
  	}
}

func (r * Region) Translate(dx int , dy int ) {
  	C.miTranslateRegion(&r.reg, C.int(dx), C.int(dy))
}

func (r * Region) Add(other Region) {
  	C.miUnion(&r.reg, &r.reg, &other.reg)
}

func (r * Region) Subtract(other Region) {
  	C.miSubtract(&r.reg, &r.reg, &other.reg)
}

func (r * Region) Intersect(other Region) {
  	C.miIntersect(&r.reg, &r.reg, &other.reg)
}

func (r * Region) Crop(rect image.Rectangle) {
	temp := NewRegionFromRect(rect)
  	r.Intersect(temp)
}

func (r * Region) IsEmpty() bool {
	if 0 == C.miRegionNotEmpty(&r.reg) {
		return true 
	}
	return false
}

func (r * Region) IsPointInside(x int , y int) bool {
  	var stubBox C.struct__Box
	if 0 != C.miPointInRegion(&r.reg, C.int(x), C.int(y), &stubBox) {
		return true
	}
  	return false
}

func (r * Region) Equals(other Region) bool {
  	if r.IsEmpty() && other.IsEmpty() {
    	return true
  	}
	if 0 != C.miRegionsEqual(&r.reg, &other.reg) {
		return true
	}
  	return false
}



func (r * Region) region_num_rects() int {
	if r.reg.data != nil {
		return int(r.reg.data.numRects)
	}
 	return 1
}



func (r * Region) GetRectVector()  (dst []image.Rectangle) {
  	boxPtr := r.region_rects()
  	numRects := r.region_num_rects()
	boxSlice := (*[1 << 30]C.BoxRec)(unsafe.Pointer(boxPtr))[:numRects:numRects]
  	for i := 0; i < numRects; i++ {
    	rect := image.Rect(int(boxSlice[i].x1), int(boxSlice[i].y1), int(boxSlice[i].x2), int(boxSlice[i].y2))
    	dst = append(dst, rect)
  	}
	return
}

func (r * Region) GetCount() int {
  	return r.region_num_rects()
}


func (r * Region) GetBounds() image.Rectangle {
  	boxPtr := C.miRegionExtents(&r.reg)
  	return image.Rect(int(boxPtr.x1), int(boxPtr.y1), int(boxPtr.x2), int(boxPtr.y2))
}

func (r * Region) region_rects() C.BoxPtr {
	if r.reg.data != nil {
 		return (C.BoxPtr)(unsafe.Pointer(uintptr(unsafe.Pointer(r.reg.data)) + C.sizeof_BoxRec))
	}
	return (C.BoxPtr)(&r.reg.extents)
}
/*
*/