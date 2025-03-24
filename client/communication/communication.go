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
func (c *CommunicationHandler) SendBatch(bets []BetInfo) error {
	serializedMsg := serializeBets(bets)
	_, err := c.conn.Write(serializedMsg)
	if err != nil {
		return err
	}
	return nil
}

// Send sends the given message through the current socket connection.
// In case of failure returns an error
func (c *CommunicationHandler) SendFinalizationMsg() error {
	serializedMsg := serializeFinalizationMsg(c.ID)
	_, err := c.conn.Write(serializedMsg)
	if err != nil {
		return err
	}
	return nil
}

// Send sends the given message through the current socket connection.
// In case of failure returns an error
func (c *CommunicationHandler) SendGetWinnersMsg() error {
	serializedMsg := serializeGetWinnersMsg(c.ID)
	_, err := c.conn.Write(serializedMsg)
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
func (c *CommunicationHandler) RecvMsg() (*BatchConfirmation, error) {
	msgType := c.recvByte()

	if msgType == BatchConfirmationMsg {
		status := c.recvByte()
		return &BatchConfirmation{Status: status}, nil
	}
	return nil, fmt.Errorf("received invalid message type %d", msgType)
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
