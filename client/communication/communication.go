package communication

import (
	"fmt"
	"io"
	"net"

	"github.com/op/go-logging"
)

type CommunicationHandler struct {
	ID            string
	conn          net.Conn
	ServerAddress string
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

// SendBatch sends the given batch to the server, and waits for the confirmation.
// In case of failure returns an error
func (c *CommunicationHandler) SendBatch(bets []BetInfo) error {
	serializedMsg := serializeBets(bets)
	// Send the message
	err := c.writeAll(serializedMsg)
	if err != nil {
		return fmt.Errorf("failed to send the message to the server. Error: %s", err)
	}
	// Receive the response
	msgType, payload, err := c.recvMsg()
	if err != nil {
		return err
	}
	// Check the status response
	if msgType == BatchConfirmationMsg {
		msg := payload.(*BatchConfirmation)
		if msg.Status == Failure {
			return fmt.Errorf("server returned failure status")
		}
	} else {
		return fmt.Errorf("expected to receive batch confirmation opcode, got %d opcode", msgType)
	}
	return nil
}

// Send sends the given message through the current socket connection.
// In case of failure returns an error
func (c *CommunicationHandler) SendFinalizationMsg() error {
	serializedMsg := serializeFinalizationMsg(c.ID)
	err := c.writeAll(serializedMsg)
	if err != nil {
		return err
	}

	return nil
}

// GetWinnersMsg sends the given message through the current socket connection,
// and returns the response received from the server.
// In case of failure returns an error
func (c *CommunicationHandler) GetWinners() (*GetWinnersResponse, error) {
	serializedMsg := serializeGetWinnersMsg(c.ID)
	err := c.writeAll(serializedMsg)
	if err != nil {
		return nil, err
	}
	msgType, payload, err := c.recvMsg()
	if err != nil {
		return nil, fmt.Errorf("failed to receive server message. Error: %s", err)
	}
	response := GetWinnersResponse{
		MessageType: msgType,
		Payload:     payload,
	}
	return &response, nil
}

// Disconnect closes the current socket connection.
// Returns an error in case of failure
func (c *CommunicationHandler) Disconnect() error {
	return c.conn.Close()
}

// RecvMsg blocks waiting for a message.
// Returns an error in case of failure
func (c *CommunicationHandler) recvMsg() (MessageType, interface{}, error) {
	msgType := c.recvByte()
	switch msgType {
	case BatchConfirmationMsg:
		status := c.recvByte()
		return msgType, &BatchConfirmation{Status: status}, nil
	case NoWinnersYetMsg:
		return msgType, nil, nil
	case WinnersMsg:
		numberOfWinners := c.recvByte()
		documents := []string{}
		// TODO: move this to a function
		for i := 0; i < int(numberOfWinners); i++ {
			documentLength := c.recvByte()
			documentBytes := c.recv(uint32(documentLength))
			document := string(documentBytes[:])
			documents = append(documents, document)
		}
		return msgType, documents, nil
	}
	err := fmt.Errorf("received invalid message type %d", msgType)
	return InvalidMsg, nil, err
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
