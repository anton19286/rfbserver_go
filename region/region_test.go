package region

import (
	"image"
	"testing"
	"math/rand"
)


func TestRegion(t *testing.T) {
	r1 := NewRegionFromRect(image.Rect(0,0,100,100))
	defer r1.Close()
	r2 := NewRegionFromRect(image.Rect(50,50,150,150))
	defer r2.Close()
	r1.Add(r2)
	n := len(r1.GetRectVector())
	if n != 3 {
		t.Errorf("rect num: %d not 3", n)
	}
	r3 := NewRegionFromRect(image.Rect(0,0,150,150))
	defer r3.Close()
	r1.Add(r3)
	n = len(r1.GetRectVector())
	if n != 1 {
		t.Errorf("rect num: %d not 1", n)
	}
	r1.Subtract(r2)
	n = len(r1.GetRectVector())
	if n != 2 {
		t.Errorf("rect num: %d not 2", n)
	}
}

func BenchmarkRegion(b *testing.B) {
	reg := NewRegion()
	r := rand.New(rand.NewSource(0))
    for i := 0; i < b.N; i++ {
		x0 := r.Intn(1000)
		y0 := r.Intn(1000)
		x1 := x0 + r.Intn(100)
		y1 := y0 + r.Intn(100)
		t := NewRegionFromRect(image.Rect(x0, y0, x1, y1))
		defer t.Close()
		reg.Add(t)
    }
	defer reg.Close()
}