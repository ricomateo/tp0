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

// Send sends the given message through the current socket connection.
// In case of failure returns an error
func (c *CommunicationHandler) Send(msg Message) error {
	serializedMsg := msg.serialize()
	err := c.writeAll(serializedMsg)
	if err != nil {
		return err
	}
	return nil
}

// Disconnect closes the current socket connection.
// Returns an error in case of failure
func (c *CommunicationHandler) Disconnect() error {
	return c.conn.Close()
}

// RecvMsg blocks waiting for a message.
// Returns an error in case of failure
func (c *CommunicationHandler) RecvMsg() (*Message, error) {
	msgType := c.recvByte()

	switch msgType {
	case ConfirmedBetMsg:
		return c.recvConfirmedBetMsg()
	}
	return nil, fmt.Errorf("invalid message type %d", msgType)
}

// recv returns `size` bytes read from the socket
func (c *CommunicationHandler) recv(size uint32) []byte {
	bytes := make([]byte, size)
	_, err := io.ReadFull(c.conn, bytes)
	if err != nil {
		log.Errorf("Failed to read connection bytes. Error: %v", err)
	}
	return bytes
}

// recvByte returns a single byte read from the socket
func (c *CommunicationHandler) recvByte() uint8 {
	return uint8(c.recv(1)[0])
}

// recvConfirmedMsg reads and deserializes a message into a ConfirmedBet message.
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

// writeAll writes all the given data to the current connection.
// In case of failure, it returns an error
func (c *CommunicationHandler) writeAll(data []byte) error {
	written := 0
	for written < len(data) {
		n, err := c.conn.Write(data[written:])

		if err != nil {
			return err
		}
		written += n
	}
	return nil
}
