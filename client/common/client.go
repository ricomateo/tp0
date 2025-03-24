package common

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	comm "github.com/7574-sistemas-distribuidos/docker-compose-init/client/communication"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// This is hardcoded here so that the tests work
// (the tests override the config.yaml)
const AGENCY_FILE = "/agency.csv"

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchSize     int
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	commHandler     comm.CommunicationHandler
	mutex           sync.Mutex
	receivedSigTerm bool
	batcher         *Batcher
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) (*Client, error) {
	batcher, err := NewBatcher(AGENCY_FILE, config.BatchSize, config.ID)
	if err != nil {
		log.Errorf("Failed to create batcher. Error: %v", err)
		return nil, err
	}
	client := &Client{
		config:          config,
		receivedSigTerm: false,
		commHandler: comm.CommunicationHandler{
			ID: config.ID,
		},
		batcher: batcher,
	}
	return client, nil
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

	// Loop until the batcher finishes reading the agency file
	for !c.batcher.Finished {
		err := c.commHandler.Connect(c.config.ServerAddress)
		if err != nil {
			log.Errorf("Failed to connect to the server. Error: %v", err)
			return
		}
		// Atomically read the SIGTERM flag
		c.mutex.Lock()
		receivedSigTerm := c.receivedSigTerm
		c.mutex.Unlock()

		// Exit in case of having received a SIGTERM signal
		if receivedSigTerm {
			c.exitGracefully()
		}

		// Get the next batch of bets
		batch := c.batcher.GetBatch()

		// Send the batch
		err = c.commHandler.SendBatch(batch)
		if err != nil {
			log.Errorf("Failed to send bet message. Error: %s", err)
		}

		// Receive response message
		msg, err := c.commHandler.RecvMsg()
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
		if msg.Status == comm.Failure {
			log.Errorf("action: batch_enviado | result: failure")
		} else {
			log.Info("action: batch_enviado | result: success")
		}

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
		c.commHandler.Disconnect()
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) exitGracefully() {
	log.Info("Shutting down the client")
	err := c.commHandler.Disconnect()
	if err != nil {
		log.Error("Failed to close connection. Error: ", err)
		os.Exit(1)
	}
	err = c.batcher.Stop()
	if err != nil {
		log.Error("Failed to stop the batcher. Error: ", err)
		os.Exit(1)
	}
	log.Info("Client exited gracefully")
	os.Exit(0)
}
