package communication

import "encoding/binary"

// serialize serializes the given message
func (m *Message) serialize() []byte {
	switch m.MessageType {
	case StoreBetMsg:
		payload := m.Payload.(BetInfo)
		return payload.serialize()
	case StoreBetBatchMsg:
		payload := m.Payload.([]BetInfo)
		return serializeBets(payload)
	}

	return []byte{}
}

// serialize serializes the BetInfo payload
func (b *BetInfo) serialize() []byte {
	messageType := byte(StoreBetMsg)
	serializedMsg := make([]byte, 0)
	serializedMsg = append(serializedMsg, messageType)

	// Serialize the fields
	serializedMsg = serializeField(serializedMsg, b.Agency)
	serializedMsg = serializeField(serializedMsg, b.Name)
	serializedMsg = serializeField(serializedMsg, b.LastName)
	serializedMsg = serializeField(serializedMsg, b.Document)
	serializedMsg = serializeField(serializedMsg, b.DateOfBirth)
	serializedMsg = serializeField(serializedMsg, b.Number)

	return serializedMsg
}

func serializeBets(bets []BetInfo) []byte {
	messageType := byte(StoreBetBatchMsg)
	serializedMsg := make([]byte, 0)
	serializedMsg = append(serializedMsg, messageType)

	// Serialize the batch size
	batchSizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(batchSizeBuf, uint32(len(bets)))
	serializedMsg = append(serializedMsg, batchSizeBuf...)

	// Serialize the fields
	for _, bet := range bets {
		serializedMsg = serializeField(serializedMsg, bet.Agency)
		serializedMsg = serializeField(serializedMsg, bet.Name)
		serializedMsg = serializeField(serializedMsg, bet.LastName)
		serializedMsg = serializeField(serializedMsg, bet.Document)
		serializedMsg = serializeField(serializedMsg, bet.DateOfBirth)
		serializedMsg = serializeField(serializedMsg, bet.Number)
	}

	return serializedMsg
}

// serializeField serializes a string field and appends it to the given buffer
func serializeField(buf []byte, field string) []byte {
	fieldLength := byte(len(field))
	buf = append(buf, fieldLength)
	buf = append(buf, []byte(field)...)
	return buf
}
