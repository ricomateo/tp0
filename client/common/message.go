package common

type MessageType = uint8

const (
	StoreBetMsg MessageType = iota
	ConfirmedBetMsg
)

type Message struct {
	messageType MessageType
	payload     interface{}
}

type ConfirmedBet struct {
	document string
	number   string
}

func StoreBetMessage(bet BetInfo) Message {
	return Message{
		messageType: StoreBetMsg,
		payload:     bet,
	}
}

func ConfirmedBetMessage(bet ConfirmedBet) Message {
	return Message{
		messageType: ConfirmedBetMsg,
		payload:     bet,
	}
}
