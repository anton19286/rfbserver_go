package rfb

import (
	"log"
	"time"
)

func (s *Server) ScreenshotUpdater() {
	var err error
	s.lastFb.Lock()
	s.lastFb.Image, err = s.screenGrabber.Grab()
	s.lastFb.Unlock()
	if err != nil {
		log.Print("server: can't grab screen")
	}
	tick := time.NewTicker(time.Second / 3)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			s.update()
		case <-s.updateRequest:
			s.update()
		case <- s.closeUpdater:
			log.Print("server: closing ScreenshotUpdater")
			return
		}
	}
}

func (s *Server) update() {
	im, err := s.screenGrabber.Grab()
	if err != nil || im == nil {
		log.Print("server: can't grab screen: ", err)
	} else {
		s.lastFb.Lock()
		s.lastFb.Image = im
		s.lastFb.Unlock()
		//	log.Printf("li.Fb grab: %v %v", im.Bounds(), len(im.Pix))
 		for c ,_ := range s.listeners {
			select {
			case c <- true:
			default:	
			}
		}
	}
}
func (s *Server) RunUpdater() {
//	s.stopUpdater()
	go s.ScreenshotUpdater()
}

func (s *Server) AddListener(c chan bool){
	if s.listeners == nil {
		s.listeners = make(map[chan bool] bool)
	}
	s.listeners[c] = true
}

func (s *Server) RemoveListener(c chan bool){
	if s.listeners == nil {
		return
	}
	delete(s.listeners, c)
	close(c)
}

func (s *Server) StopUpdater() {
	close(s.closeUpdater)
}
