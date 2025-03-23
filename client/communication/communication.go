package communication

import (
	"fmt"
	"io"
	"net"

	"github.com/op/go-logging"
)

type CommunicationHandler struct {
	ID   string
	conn net.Conn
}

var log = logging.MustGetLogger("log")

// Connect Initializes client socket, connecting to the given address.
// In case of failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *CommunicationHandler) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *CommunicationHandler) Send(msg Message) error {
	serializedMsg := msg.serialize()
	n, err := c.conn.Write(serializedMsg)
	if err != nil {
		return err
	}
	if n < len(serializedMsg) {
		return fmt.Errorf("expected to send %v bytes but sent %v", len(serializedMsg), n)
	}
	return nil
}

func (c *CommunicationHandler) Disconnect() error {
	return c.conn.Close()
}

func (c *CommunicationHandler) RecvMsg() (*Message, error) {
	msgType := c.recvByte()

	switch msgType {
	case ConfirmedBetMsg:
		return c.recvConfirmedBetMsg()
	}
	return nil, fmt.Errorf("invalid message type %d", msgType)
}

func (c *CommunicationHandler) recv(size uint32) []byte {
	bytes := make([]byte, size)
	_, err := io.ReadFull(c.conn, bytes)
	if err != nil {
		log.Errorf("Failed to read connection bytes. Error: %v", err)
	}
	return bytes
}

func (c *CommunicationHandler) recvByte() uint8 {
	return uint8(c.recv(1)[0])
}

func (c *CommunicationHandler) recvConfirmedBetMsg() (*Message, error) {
	// Decode document
	documentLen := uint32(c.recvByte())
	document := string(c.recv(documentLen)[:])
	// Decode number
	numberLen := uint32(c.recvByte())
	number := string(c.recv(numberLen)[:])

	payload := ConfirmedBet{
		Document: document,
		Number:   number,
	}
	msg := ConfirmedBetMessage(payload)
	return &msg, nil
}
