package common

import (
	"bufio"
	"os"
	"strings"

	comm "github.com/7574-sistemas-distribuidos/docker-compose-init/client/communication"
)

// Batcher holds the data and logic for batching the bets
type Batcher struct {
	Finished    bool
	file        *os.File
	fileScanner *bufio.Scanner
	batchSize   int
	agencyId    string
}

// NewBatcher creates a new batcher that reads the records from the `agencyFile`.
// Returns an error if it fails to create the batcher.
func NewBatcher(agencyFile string, batchSize int, agencyId string) (*Batcher, error) {
	file, err := os.Open(agencyFile)
	if err != nil {
		return nil, err
	}

	fileScanner := bufio.NewScanner(file)
	batcher := Batcher{
		Finished:    false,
		file:        file,
		fileScanner: fileScanner,
		batchSize:   batchSize,
		agencyId:    agencyId,
	}
	return &batcher, nil
}

// GetBatch returns the next batch of at most `batchSize` size.
func (b *Batcher) GetBatch() []comm.BetInfo {
	bets := []comm.BetInfo{}
	for i := 0; i < b.batchSize; i++ {
		b.Finished = !b.fileScanner.Scan()
		if b.Finished {
			break
		}
		line := b.fileScanner.Text()
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
			Agency:      b.agencyId,
			Name:        name,
			LastName:    lastName,
			Document:    document,
			DateOfBirth: birthdate,
			Number:      number,
		}
		bets = append(bets, bet)
	}
	return bets
}

// Stop closes the underlying bets file.
func (b *Batcher) Stop() error {
	log.Info("Closing the agency bets file")
	err := b.file.Close()
	return err
}
