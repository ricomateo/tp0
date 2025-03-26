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

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchSize     int
	AgencyFile    string
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
	batcher, err := NewBatcher(config.AgencyFile, config.BatchSize, config.ID)
	if err != nil {
		log.Errorf("Failed to create batcher. Error: %v", err)
		return nil, err
	}
	client := &Client{
		config:          config,
		receivedSigTerm: false,
		commHandler: comm.CommunicationHandler{
			ID:            config.ID,
			ServerAddress: config.ServerAddress,
		},
		batcher: batcher,
	}
	return client, nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	defer c.checkForShutdown()
	c.setSigTermHandler()

	// Loop until the batcher finishes reading the agency file
	for !c.batcher.Finished {
		c.checkForShutdown()

		// Get the next batch of bets
		batch := c.batcher.GetBatch()

		// Send the batch
		err := c.commHandler.SendBatch(batch)
		if err != nil {
			log.Errorf("action: batch_enviado | result: failure | error: %s", err)
		} else {
			log.Info("action: batch_enviado | result: success")
		}

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}

	err := c.commHandler.SendFinalizationMsg()
	if err != nil {
		log.Error("action: finalization_enviado | result: failure")
		return
	}
	log.Info("action: finalization_enviado | result: success")

	betWinners, err := c.requestWinners()
	if err != nil {
		log.Errorf("action: consulta_ganadores | result: failure | error: %s", err)
		return
	}
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(betWinners))
}

// requestWinners loops requesting the winners to the server, until it responds with the winners
// It sleeps for a second between requests
func (c *Client) requestWinners() ([]string, error) {
	for {
		c.checkForShutdown()
		response, err := c.commHandler.GetWinners()
		if err != nil {
			return nil, err
		}
		// If there are no winners yet, sleep some time and then request the winners again
		if response.MessageType == comm.NoWinnersYetMsg {
			time.Sleep(1 * time.Second)
			continue
		}
		// If the server responds with the winners, break
		if response.MessageType == comm.WinnersMsg {
			winners := response.Payload.([]string)
			return winners, nil
		}
	}
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

func (c *Client) setSigTermHandler() {
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
}

func (c *Client) checkForShutdown() {
	// Atomically read the SIGTERM flag
	c.mutex.Lock()
	receivedSigTerm := c.receivedSigTerm
	c.mutex.Unlock()

	// Exit in case of having received a SIGTERM signal
	if receivedSigTerm {
		c.exitGracefully()
	}
}
