package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type Server struct {
	port int
}

func New(port int) *Server {
	s := Server{port}
	return &s
}

func (s Server) Run() {

	server, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if server == nil {
		panic("couldn't start listening: " + err.Error())
	}

	conn := clientConn(server)
	for {
		go handleConn(<-conn)
	}
}

func clientConn(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn)
	i := 0

	go func() {
		for {
			client, err := listener.Accept()
			if client == nil {
				fmt.Printf("couldn't accept: " + err.Error())
				continue
			}

			i++
			log.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())

			ch <- client
		}
	}()

	return ch
}

func handleConn(client net.Conn) {
	defer func() {
		if err := client.Close(); err != nil {
			log.Println("method", "close", "err", err)
		}
		log.Printf("disconnected: %v\n", client.RemoteAddr())
	}()

	// echo
	if _, err := io.Copy(client, client); err != nil {
		log.Println("method", "echo", "err", err)
	}
}
