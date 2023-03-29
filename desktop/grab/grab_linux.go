package grab

import (
	"image"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type ScreenGrabber interface {
	Bounds() image.Rectangle
	Grab() (*image.RGBA, error)
}

type X11ScreenGrabber struct {
	c *xgb.Conn
}

func NewScreenGrabber() (*X11ScreenGrabber, error) {
	c, err := xgb.NewConn()
	return &X11ScreenGrabber{c}, err
}

func (g *X11ScreenGrabber) Bounds() image.Rectangle {
	screen := xproto.Setup(g.c).DefaultScreen(g.c)
	x := int(screen.WidthInPixels)
	y := int(screen.HeightInPixels)
	return image.Rect(0, 0, x, y)
}

func (g *X11ScreenGrabber) Grab() (img *image.RGBA, err error) {
	drawable := xproto.Drawable(xproto.Setup(g.c).DefaultScreen(g.c).Root)
	rect := img.Bounds()
	x0, y0 := int16(rect.Min.X), int16(rect.Min.Y)
	dx, dy := uint16(rect.Dx()), uint16(rect.Dy())
	const planemask = 0xffffffff
	const format = xproto.ImageFormatZPixmap

	im, err := xproto.GetImage(g.c, format, drawable, x0, y0, dx, dy, planemask).Reply()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(im.Data); i += 4 {
		img.Pix[i], img.Pix[i+2], img.Pix[i+1], img.Pix[i+3] = 
		im.Data[i+2], im.Data[i], im.Data[i+1], 255
	}
	
	return img, nil
}
