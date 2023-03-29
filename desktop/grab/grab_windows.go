package grab

import (
	"fmt"
	"image"
	"github.com/vova616/screenshot"
)

type ScreenGrabber interface {
	Bounds() image.Rectangle
	Grab() (*image.RGBA, error)
}

type WINScreenGrabber struct{}

func NewScreenGrabber() (*WINScreenGrabber, error) {
	return &WINScreenGrabber{}, nil
}

func (g *WINScreenGrabber) Bounds() image.Rectangle {
	r, _ := screenshot.ScreenRect()
	return r
}

func (g *WINScreenGrabber) Grab() (m *image.RGBA, err error) {
    defer func() {
	    if r := recover(); r != nil {
    	    err = fmt.Errorf("recovered in Grab: %v", r)
    	}
	}()
	m, err = screenshot.CaptureScreen()
	return 
}
