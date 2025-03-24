package communication

type MessageType = uint8

// The different types of messages
const (
	StoreBetBatchMsg MessageType = iota
	BatchConfirmationMsg
	FinalizationMsg
	GetWinnersMsg
)

type BatchStatus = uint8

// The different types of messages
const (
	Failure BatchStatus = iota
	Success
)

type BetInfo struct {
	Agency      string
	Name        string
	LastName    string
	Document    string
	DateOfBirth string
	Number      string
}

type BatchConfirmation struct {
	Status BatchStatus
}
