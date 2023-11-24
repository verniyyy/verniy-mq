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
	"github.com/verniyyy/verniy-mq/src/util"
)

const (
	_ = iota
	QuitCMD
	PingCMD
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
	sessID := util.GenULID()
	defer func() {
		log.Printf("connection closed: %v, session %s", conn.RemoteAddr(), sessID)
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	log.Printf("connection established on %v, session %s", conn.RemoteAddr(), sessID)

	r := bufio.NewReader(conn)
	for {
		header, err := readHeader(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			return
		}
		if header.Command == QuitCMD {
			break
		}

		log.Printf("header: %+v\n", header)

		app := src.NewMessageQueueApplication(h.mqManager)
		if err := func() error {
			switch header.Command {
			case PingCMD:
				log.Println("ping")
				return nil
			case CreateQueueCMD:
				return app.CreateQueue(context.Background(), header.accountIDString(), header.queueNameString())
			case DeleteQueueCMD:
				return app.DeleteQueue(context.Background(), header.accountIDString(), header.queueNameString())
			case PublishCMD:
				data, err := io.ReadAll(r)
				if err != nil {
					return err
				}
				return app.Publish(context.Background(), header.accountIDString(), header.queueNameString(), data)
			case ConsumeCMD:
				m, err := app.Consume(context.Background(), header.accountIDString(), header.queueNameString())
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
				return app.Delete(context.Background(), header.accountIDString(), string(header.QueueName[:]), string(id[:]))
			default:
				if header.isBlank() {
					return nil
				}
				return fmt.Errorf("invalid command: %v", header.Command)
			}
		}(); err != nil {
			log.Printf("error: %v\n", err)
		}

		conn.Write([]byte("Message received.\n"))
	}
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
	Command   uint8
	AccountID [32]rune
	QueueName [128]rune
}

// String implements of fmt.Stringer
func (h HeaderField) String() string {
	return fmt.Sprintf("{Command:%d, AccountID:%s, QueueName:%s}",
		h.Command, h.accountIDString(), h.queueNameString())
}

// isBlank ...
func (h HeaderField) isBlank() bool {
	var blank HeaderField
	return h == blank
}

// accountIDString ...
func (h HeaderField) accountIDString() string {
	const accountIDTest = "01HG17X22440GTQW3AS6WHCF0K" // ulid
	accountID := string(h.AccountID[:])
	if accountID == "" {
		return accountIDTest
	}
	return accountID
}

// queueNameString ...
func (h HeaderField) queueNameString() string {
	return string(h.QueueName[:])
}

// MessageID ...
type MessageID [src.MessageIDSize]byte

// readHeader ...
func readHeader(r io.Reader) (HeaderField, error) {
	buf := make([]byte, headerFieldSize)
	received, err := r.Read(buf)
	if err != nil {
		return HeaderField{}, err
	}

	var headerField HeaderField
	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &headerField); err != nil {
		log.Printf("received: %v\n", received)
		return HeaderField{}, err
	}

	return headerField, nil
}
