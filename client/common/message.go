package common

type MessageType = uint8

const (
	StoreBet MessageType = iota
	ConfirmedBet
)

type Message struct {
	messageType MessageType
	payload     interface{}
}

func StoreBetMessage(bet BetInfo) Message {
	return Message{
		messageType: StoreBet,
		payload:     bet,
	}
}
