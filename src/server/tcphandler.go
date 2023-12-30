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
	"os"
	"strconv"
	"strings"

	"github.com/verniyyy/verniy-mq/src"
	"github.com/verniyyy/verniy-mq/src/util"
)

const (
	_ = iota
	QuitCMD
	PingCMD
	CreateQueueCMD
	ListQueueCMD
	DeleteQueueCMD
	PublishCMD
	ConsumeCMD
	DeleteCMD
)

const (
	_ uint8 = iota
	OK
	Error
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
	connID := util.GenULID()
	defer func() {
		log.Printf("connection closed: %v, connection %s", conn.RemoteAddr(), connID)
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	log.Printf("connection established on %v, connection %s", conn.RemoteAddr(), connID)

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	authField, err := read[AuthField](r, authFieldSize)
	if err != nil {
		log.Println(err)
		return
	}
	if !auth(authField) {
		fmt.Printf("authField: %v\n", authField)
		log.Println("authentication failed")
		return
	}

	sessID := NewSessionID(util.GenULID)
	buf := new(bytes.Buffer)
	if err := binary.Write(
		buf,
		binary.BigEndian,
		sessID,
	); err != nil {
		log.Println(err)
		return
	}

	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
		return
	}
	if err := w.Flush(); err != nil {
		log.Println(err)
		return
	}

	log.Printf("auth ok session id: %v\n", sessID)

	for {
		header, err := read[HeaderField](r, headerFieldSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			return
		}
		if header.SessionID != sessID {
			log.Println("session id mismatch")
			log.Printf("header.sessionIDString(): %p\n", []byte(header.String()))
			log.Printf("sessID.String(): %p\n", []byte(sessID.String()))
			res, err := NewResponse(Error, []byte("session timeout")).encode()
			if err != nil {
				log.Printf("error: %v\n", err)
			}

			if _, err := writeWithFlush(w, res); err != nil {
				log.Printf("error: %v\n", err)
			}
			break
		}
		if header.Command == QuitCMD {
			break
		}

		log.Printf("header: %+v\n", header)

		app := src.NewMessageQueueApplication(h.mqManager)
		resData, err := func() ([]byte, error) {
			switch header.Command {
			case PingCMD:
				log.Println("PingCMD")
				return []byte("pong"), nil
			case CreateQueueCMD:
				log.Println("CreateQueueCMD")
				if err := app.CreateQueue(context.Background(), authField.accountIDString(), header.queueNameString()); err != nil {
					return nil, err
				}
				return nil, nil
			case ListQueueCMD:
				log.Println("list queue cmd")
				return nil, nil
			case DeleteQueueCMD:
				log.Println("DeleteQueueCMD")
				if err := app.DeleteQueue(context.Background(), authField.accountIDString(), header.queueNameString()); err != nil {
					return nil, err
				}
				return nil, nil
			case PublishCMD:
				log.Println("PublishCMD")
				data := make([]byte, header.DataSize)
				_, err := r.Read(data)
				if err != nil {
					return nil, err
				}
				fmt.Printf("data: %v\n", data)
				return nil, app.Publish(context.Background(), authField.accountIDString(), header.queueNameString(), data)
			case ConsumeCMD:
				log.Println("ConsumeCMD")
				m, err := app.Consume(context.Background(), authField.accountIDString(), header.queueNameString())
				if err != nil {
					return nil, err
				}
				return m.Bytes(), nil
			case DeleteCMD:
				log.Println("DeleteCMD")
				var id MessageID
				_, err := r.Read(id[:])
				if err != nil && err != io.EOF {
					return nil, err
				}

				log.Printf("delete message id: %v\n", string(id[:]))
				return nil, app.Delete(context.Background(), authField.accountIDString(), header.queueNameString(), string(id[:]))
			default:
				log.Println("invalid cmd")
				if header.isBlank() {
					return nil, nil
				}
				return nil, fmt.Errorf("invalid command: %v", header.Command)
			}
		}()
		res, err := func() ([]byte, error) {
			if err != nil {
				log.Printf("error: %v\n", err)
				return NewResponse(Error, []byte(err.Error())).encode()
			}
			return NewResponse(OK, resData).encode()
		}()
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}

		n, err := writeWithFlush(w, res)
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		log.Printf("response write n: %v\n", n)
	}
}

type bufWriter interface {
	Write([]byte) (int, error)
	Flush() error
}

func writeWithFlush(w bufWriter, p []byte) (int, error) {
	n, err := w.Write(p)
	if err := w.Flush(); err != nil {
		return n, err
	}
	return n, err
}

// SessionID ...
type SessionID [32]rune

// String ...
func (sid SessionID) String() string {
	return string(sid[:])
}

// SessionIDGenerator ...
type SessionIDGenerator func() string

// NewSessionID ...
func NewSessionID(g SessionIDGenerator) SessionID {
	var sessID SessionID
	copy(sessID[:], []rune(g()))
	return sessID
}

const (
	queueNameStrSize = 128
	ByteSizeOfRune   = 4
	headerFieldSize  = accountIDStrSize*ByteSizeOfRune +
		1 + // Command field size
		queueNameStrSize*ByteSizeOfRune +
		8 // data size field
)

func init() {
	fmt.Printf("headerFieldSize: %v\n", headerFieldSize)
}

// HeaderField ...
type HeaderField struct {
	SessionID SessionID
	Command   uint8
	QueueName [128]rune
	DataSize  uint64
}

// String implements of fmt.Stringer
func (h HeaderField) String() string {
	return fmt.Sprintf("{SessionID:%s, Command:%d, QueueName:%s}",
		h.SessionID.String(), h.Command, h.queueNameString())
}

// isBlank ...
func (h HeaderField) isBlank() bool {
	var blank HeaderField
	return h == blank
}

// queueNameString ...
func (h HeaderField) queueNameString() string {
	return strings.Replace(string(h.QueueName[:]), "\x00", "", -1)
}

// MessageID ...
type MessageID [src.MessageIDSize]byte

// read ...
func read[T any](r io.Reader, bufSize uint64) (T, error) {
	var v T
	buf := make([]byte, bufSize)
	received, err := r.Read(buf)
	if err != nil {
		return v, err
	}

	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &v); err != nil {
		log.Printf("received: %v\n", received)
		return v, err
	}

	return v, nil
}

// auth ...
func auth(a AuthField) bool {
	if disableAuth, _ := strconv.ParseBool(os.Getenv("DISABLE_AUTH")); disableAuth {
		return true
	}

	type Users map[string]string
	users := make(Users)
	users[accountIDTest] = passwordTest

	if _, ok := users[a.accountIDString()]; !ok {
		return false
	}
	if users[a.accountIDString()] != a.passwordString() {
		log.Println("invalid password")
		return false
	}

	// Check if the username and password are valid
	// if _, ok := users[a.accountIDString()]; !ok || users[a.accountIDString()] != a.passwordString() {
	// 	// The username or password is invalid
	// 	return false
	// }

	return true
}

const (
	accountIDStrSize = 32
	passwordStrSize  = 64
	authFieldSize    = accountIDStrSize*ByteSizeOfRune +
		passwordStrSize*ByteSizeOfRune
)

const (
	accountIDTest = "01HG17X22440GTQW3AS6WHCF0K" // ulid
	passwordTest  = "P@ssw0rd"
)

// AuthField ...
type AuthField struct {
	AccountID [32]rune
	Password  [64]rune
}

// String ...
func (a AuthField) String() string {
	return fmt.Sprintf("{AccountID:%s, Password:%s}",
		a.accountIDString(), a.passwordString())
}

// accountIDString ...
func (a AuthField) accountIDString() string {
	accountID := string(a.AccountID[:])
	if accountID == "" {
		return accountIDTest
	}
	return util.TrimNullChar(accountID)
}

// passwordString ...
func (a AuthField) passwordString() string {
	password := string(a.Password[:])
	return util.TrimNullChar(password)
}

// Response ...
type Response struct {
	HeaderField struct {
		Result   uint8
		DataSize uint64
	}
	Data []byte
}

// NewResponse ...
func NewResponse(result uint8, data []byte) Response {
	res := Response{}
	res.HeaderField.Result = result
	res.HeaderField.DataSize = uint64(len(data))
	res.Data = data
	return res
}

// encode ...
func (res Response) encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(
		buf,
		binary.BigEndian,
		res.HeaderField,
	); err != nil {
		return []byte{}, err
	}
	if res.Data == nil {
		return buf.Bytes(), nil
	}

	return append(buf.Bytes(), res.Data...), nil
}
