package rfb

import (
	"image"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"rfbserver/desktop/grab"
	"rfbserver/rfb/encoder"
	"sync"
	"time"

	//	"rfbserver/desktop/input"

	"golang.org/x/net/websocket"
)

type ServerConfig struct {
	Width           uint16
	Height          uint16
	PixelFormat     encoder.PixelFormat
	DesktopName     string
	clientEncodings []int32
}

type messageHandler struct {
	Id      byte
	Handler func([]byte, *Client)
	Name    string
}

type LockableImage struct {
	sync.RWMutex
	Image *image.RGBA
}

type Server struct {
	muc    sync.RWMutex
	Config ServerConfig
	lastFb *LockableImage
	//	userInput              input.UserInput
	screenGrabber grab.ScreenGrabber
	updateRequest chan bool
	closeUpdater  chan bool
	stat          chan Stat
	listeners     map[chan bool]bool
}

func NewServer() *Server {
	s := &Server{}
	s.Config = ServerConfig{DesktopName: "go rfb server",
		PixelFormat: encoder.DefaultPixelFormat}

	var err error
	s.screenGrabber, err = grab.NewScreenGrabber()
	if err != nil {
		log.Fatal("server: error connectig screen grabber")
	}
	//	s.userInput, err = input.NewUserInput()
	//	if err != nil {
	//		log.Fatal("server: error connectig user input")
	//	}
	r := s.screenGrabber.Bounds()
	log.Printf("server: screen size: %v", r)
	s.Config.Width, s.Config.Height = uint16(r.Dx()), uint16(r.Dy())
	s.lastFb = &LockableImage{Image: image.NewRGBA(r)}
	s.updateRequest = make(chan bool)
	s.closeUpdater = make(chan bool)
	s.listeners = make(map[chan bool]bool)
	s.stat = make(chan Stat)
	go s.stst()
	// TODO add waitgroup?
	//	defer close(s.stat)
	return s
}

type Stat struct {
	t     time.Duration
	size1 int
	size2 int
}

func (this *Server) stst() {
	sum := func(a []Stat) Stat {
		s := Stat{}
		for _, v := range a {
			s.t += v.t
			s.size1 += v.size1
			s.size2 += v.size2
		}
		return s
	}
	s := []Stat{}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if len(s) != 0 {
				avg := int(sum(s).t/time.Millisecond) / len(s)
				log.Printf("server stat: %v msg, t_avg :%v ms, size loss: %v, size lossless: %v",
					len(s), avg, sum(s).size1, sum(s).size2)
			}
			s = s[:0]
		case v, ok := <-this.stat:
			if !ok {
				return
			}
			s = append(s, v)
		}
	}
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		c := s.addClient(conn)
		log.Printf("client connected to: %v from: %v",
			conn.LocalAddr(),
			conn.RemoteAddr())
		go c.serve()
	}
}

func (s *Server) websocketHandler(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	c := s.addClient(conn)
	c.serve()
}

// checkOrigin replaces default handshake handler as
// Firefox can send null in Origin field
func checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if config.Origin == nil {
		config.Origin = &url.URL{Host: "null"}
		err = nil
	}
	return err
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ws := websocket.Server{
		Handler:   s.websocketHandler,
		Handshake: checkOrigin}
	ws.ServeHTTP(w, req)
}

func (s *Server) addClient(conn io.ReadWriteCloser) *Client {
	// TODO Add waitgroup for updater
	s.RunUpdater()
	listener := make(chan bool, 1)
	s.AddListener(listener)
	bounds := s.lastFb.Image.Bounds()
	c := NewClient(s, conn, bounds, listener)
	return c
}
