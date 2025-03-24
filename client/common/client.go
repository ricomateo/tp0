package common

import (
	"bufio"
	"os"
	"os/signal"
	"strings"
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
	BetInfo       comm.BetInfo
	BatchSize     int
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	commHandler     comm.CommunicationHandler
	mutex           sync.Mutex
	receivedSigTerm bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:          config,
		receivedSigTerm: false,
		commHandler: comm.CommunicationHandler{
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

	file, err := os.Open("/agency.csv")
	if err != nil {
		log.Errorf("Failed to open agency file: %v", err)
		return
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	// Scan() returns false when it gets to the end of the file
	for fileScanner.Scan() {
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
		batch := c.getBatch(fileScanner)
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

func (c *Client) getBatch(fileScanner *bufio.Scanner) []comm.BetInfo {
	bets := []comm.BetInfo{}
	for i := 0; i < c.config.BatchSize; i++ {
		line := fileScanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			log.Error("Failed to parse bet record. Error: missing fields")
			continue
		}
		name := fields[0]
		lastName := fields[1]
		document := fields[2]
		birthdate := fields[3]
		number := fields[4]
		bet := comm.BetInfo{
			Agency:      c.config.ID,
			Name:        name,
			LastName:    lastName,
			Document:    document,
			DateOfBirth: birthdate,
			Number:      number,
		}
		bets = append(bets, bet)
		// if Scan() return false it means there is no more data to read
		if !fileScanner.Scan() {
			break
		}
	}
	return bets
}

func (c *Client) exitGracefully() {
	log.Info("Shutting down socket connection")
	err := c.commHandler.Disconnect()
	if err != nil {
		log.Error("Failed to close connection. Error: ", err)
		os.Exit(1)
	}
	log.Info("Client exited gracefully")
	os.Exit(0)
}
