package common

import (
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
	commHandler     CommunicationHandler
	mutex           sync.Mutex
	receivedSigTerm bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:          config,
		receivedSigTerm: false,
		commHandler: CommunicationHandler{
			ID: config.ID,
		},
	}
	return client
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
		c.commHandler.connect(c.config.ServerAddress)

		// Atomically read the SIGTERM flag
		c.mutex.Lock()
		receivedSigTerm := c.receivedSigTerm
		c.mutex.Unlock()

		// Exit in case of having received a SIGTERM signal
		if receivedSigTerm {
			c.exitGracefully()
		}
		betMsg := c.config.BetInfo.serialize()
		err := c.commHandler.send(betMsg)
		if err != nil {
			log.Errorf("Failed to send bet message. Error: %s", err)
		}

		msg, err := c.commHandler.recv_msg()
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		// TODO: check if this log can be removed
		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", msg.document, msg.number)

		// TODO: add commHandler method for closing
		c.commHandler.conn.Close()

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) exitGracefully() {
	log.Info("Shutting down socket connection")
	err := c.commHandler.conn.Close()
	if err != nil {
		log.Error("Failed to close connection. Error: ", err)
		os.Exit(1)
	}
	log.Info("Client exited gracefully")
	os.Exit(0)
}
