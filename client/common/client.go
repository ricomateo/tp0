package common

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type BetInfo struct {
	Name        string
	LastName    string
	Document    string
	DateOfBirth string
	Number      string
}

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BetInfo       BetInfo
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	conn            net.Conn
	mutex           sync.Mutex // Mutex to synchronize the receivedSigTerm flag read/write
	receivedSigTerm bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:          config,
		receivedSigTerm: false,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	signalChannel := make(chan os.Signal, 2)
	// Set the SIGTERM handler
	signal.Notify(signalChannel, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		if sig == syscall.SIGTERM {
			log.Info("Received SIGTERM signal")
			c.mutex.Lock()
			c.receivedSigTerm = true
			c.mutex.Unlock()
		}
	}()
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// Atomically read the SIGTERM flag
		c.mutex.Lock()
		receivedSigTerm := c.receivedSigTerm
		c.mutex.Unlock()

		// Exit in case of having received a SIGTERM signal
		if receivedSigTerm {
			c.exitGracefully()
		}
		betMsg := c.config.BetInfo.serialize()
		log.Info("Serialized bet info: %v", betMsg)
		n, err := c.conn.Write(betMsg)
		if err != nil {
			log.Error("Failed to write bet message. Error: %v", err)
		}
		if n < len(betMsg) {
			log.Info("Expected to write %v bytes, wrote %v bytes", len(betMsg), n)
		}

		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) exitGracefully() {
	log.Info("Shutting down socket connection")
	err := c.conn.Close()
	if err != nil {
		log.Error("Failed to close connection. Error: ", err)
		os.Exit(1)
	}
	log.Info("Client exited gracefully")
	os.Exit(0)
}
