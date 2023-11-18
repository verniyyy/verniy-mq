package src

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

// Server ...
type Server interface {
	Run()
}

// NewServer ...
func NewServer(host string, port int) Server {
	return server{
		messageQueueApplication: NewMessageQueueApplication(),
		host:                    host,
		port:                    fmt.Sprint(port),
	}
}

// server ...
type server struct {
	messageQueueApplication MessageQueueApplication
	host                    string
	port                    string
}

// Run ...
func (s server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.host, s.port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Printf("listening on %s:%s", s.host, s.port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleRequest(conn, s.messageQueueApplication)
	}
}

const (
	_ = iota
	CreateQueueCMD
	DeleteQueueCMD
	PublishCMD
	ConsumeCMD
	DeleteCMD
)

// handleRequest ...
func handleRequest(conn net.Conn, app MessageQueueApplication) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	r := bufio.NewReader(conn)

	header, err := readHeader(r)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("header: %+v\n", header)

	if err := func() error {
		switch header.Command {
		case CreateQueueCMD:
			return app.CreateQueue(context.Background(), string(header.QueueName[:]))
		case DeleteQueueCMD:
			return app.DeleteQueue(context.Background(), string(header.QueueName[:]))
		case PublishCMD:
			return app.Publish(context.Background(), string(header.QueueName[:]), r)
		case ConsumeCMD:
			m, err := app.Consume(context.Background(), string(header.QueueName[:]), r)
			if err != nil {
				return err
			}
			conn.Write(m.Bytes())
			return nil
		case DeleteCMD:
			return app.Delete(context.Background(), string(header.QueueName[:]))
		default:
			return fmt.Errorf("invalid command: %v", header.Command)
		}
	}(); err != nil {
		log.Printf("error: %v\n", err)
	}

	conn.Write([]byte("Message received.\n"))
}

// headerFieldSize ...
const headerFieldSize = 161

// HeaderField ...
type HeaderField struct {
	AccountID [32]rune
	QueueName [128]rune
	Command   uint8
}

// readHeader ...
func readHeader(r io.Reader) (HeaderField, error) {
	buf := make([]byte, headerFieldSize)
	received, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return HeaderField{}, err
	}
	if received != headerFieldSize {
		return HeaderField{}, fmt.Errorf("invalid buf size: %v", received)
	}

	var headerField HeaderField
	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &headerField); err != nil {
		return HeaderField{}, err
	}

	return headerField, nil
}
