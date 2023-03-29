package main

import (
	"net"
//	"net/http"
	"rfbserver/rfb"
//	"github.com/GeertJohan/go.rice"
)


func main() {
	Server := rfb.NewServer()

//	http.Handle("/websockify", Server)
//	http.Handle("/", http.FileServer(rice.MustFindBox("http-files").HTTPBox()))
//	err := http.ListenAndServe(":8080", nil)
//	certPath := "./localhost/"
//	err := http.ListenAndServeTLS(":443", certPath+"server.pem", certPath+"privkey.pem", nil)

//	if err != nil {
//        	panic("ListenAndServe: " + err.Error())
//	}

	l, err := net.Listen("tcp", ":5900")
	if err != nil {
		panic("Listen: " + err.Error())
	}
	defer l.Close()
	Server.Serve(l) 

}
