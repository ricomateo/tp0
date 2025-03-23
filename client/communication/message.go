package communication

type MessageType = uint8

// The different types of messages
const (
	StoreBetMsg MessageType = iota
	ConfirmedBetMsg
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

func ConfirmedBetMessage(bet ConfirmedBet) Message {
	return Message{
		MessageType: ConfirmedBetMsg,
		Payload:     bet,
	}
}
