package communication

type MessageType = uint8

// The different types of messages
const (
	StoreBetMsg MessageType = iota
	ConfirmedBetMsg
	StoreBetBatchMsg
)

// Message struct containing the message type and its payload
type Message struct {
	MessageType MessageType
	Payload     interface{}
}

type ConfirmedBet struct {
	Document string
	Number   string
}

type BetInfo struct {
	Agency      string
	Name        string
	LastName    string
	Document    string
	DateOfBirth string
	Number      string
}

func StoreBetMessage(bet BetInfo) Message {
	return Message{
		MessageType: StoreBetMsg,
		Payload:     bet,
	}
}

func StoreBetBatchMessage(bets []BetInfo) Message {
	return Message{
		MessageType: StoreBetBatchMsg,
		Payload:     bets,
	}
}

func ConfirmedBetMessage(bet ConfirmedBet) Message {
	return Message{
		MessageType: ConfirmedBetMsg,
		Payload:     bet,
	}
}
