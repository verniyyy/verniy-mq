package server

import (
	"fmt"
	"log"
	"net"

	"github.com/verniyyy/verniy-mq/src"
)

// NewTCPServer ...
func NewTCPServer(host string, port int, mqm src.MQManager) Server {
	return tcpServer{
		host:    host,
		port:    fmt.Sprint(port),
		handler: newTCPHandler(mqm),
	}
}

// tcpServer ...
type tcpServer struct {
	host    string
	port    string
	handler TCPHandler
}

// Run ...
func (s tcpServer) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.host, s.port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Printf("listening on %s:%s by tcp", s.host, s.port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.handler.HandleRequest(conn)
	}
}
