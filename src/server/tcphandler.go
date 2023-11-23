package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/verniyyy/verniy-mq/src"
)

const (
	PingCMD = iota
	CreateQueueCMD
	DeleteQueueCMD
	PublishCMD
	ConsumeCMD
	DeleteCMD
)

// TCPHandler ...
type TCPHandler interface {
	HandleRequest(net.Conn)
}

// newTCPHandler ...
func newTCPHandler(mqm src.MQManager) TCPHandler {
	return tcpHandler{
		mqManager: mqm,
	}
}

// tcpHandler ...
type tcpHandler struct {
	mqManager src.MQManager
}

// HandleRequest ...
func (h tcpHandler) HandleRequest(conn net.Conn) {
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

	app := src.NewMessageQueueApplication(h.mqManager)
	if err := func() error {
		switch header.Command {
		case PingCMD:
			log.Println("ping")
			return nil
		case CreateQueueCMD:
			return app.CreateQueue(context.Background(), string(header.AccountID[:]), string(header.QueueName[:]))
		case DeleteQueueCMD:
			return app.DeleteQueue(context.Background(), string(header.AccountID[:]), string(header.QueueName[:]))
		case PublishCMD:
			return app.Publish(context.Background(), string(header.AccountID[:]), string(header.QueueName[:]), r)
		case ConsumeCMD:
			m, err := app.Consume(context.Background(), string(header.AccountID[:]), string(header.QueueName[:]))
			if err != nil {
				return err
			}
			conn.Write(m.Bytes())
			return nil
		case DeleteCMD:
			var id MessageID
			_, err := r.Read(id[:])
			if err != nil && err != io.EOF {
				return err
			}
			return app.Delete(context.Background(), string(header.AccountID[:]), string(header.QueueName[:]), string(id[:]))
		default:
			return fmt.Errorf("invalid command: %v", header.Command)
		}
	}(); err != nil {
		log.Printf("error: %v\n", err)
	}

	conn.Write([]byte("Message received.\n"))
}

const (
	accountIDStrSize = 32
	queueNameStrSize = 128
	ByteSizeOfRune   = 4
	headerFieldSize  = accountIDStrSize*ByteSizeOfRune +
		queueNameStrSize*ByteSizeOfRune +
		1 // Command field size
)

func init() {
	fmt.Printf("headerFieldSize: %v\n", headerFieldSize)
}

// HeaderField ...
type HeaderField struct {
	AccountID [32]rune
	QueueName [128]rune
	Command   uint8
}

func (h HeaderField) String() string {
	return fmt.Sprintf("{AccountID:%s, QueueName:%s, Command:%d}",
		h.AccountIDString(), h.QueueNameString(), h.Command)
}

func (h HeaderField) AccountIDString() string {
	return string(h.AccountID[:])
}

func (h HeaderField) QueueNameString() string {
	return string(h.QueueName[:])
}

// MessageID ...
type MessageID [src.MessageIDSize]byte

// readHeader ...
func readHeader(r io.Reader) (HeaderField, error) {
	buf := make([]byte, headerFieldSize)
	received, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return HeaderField{}, err
	}
	// if received != headerFieldSize {
	// 	return HeaderField{}, fmt.Errorf("invalid buf size: %v", received)
	// }

	var headerField HeaderField
	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &headerField); err != nil {
		fmt.Printf("received: %v\n", received)
		return HeaderField{}, err
	}

	return headerField, nil
}
