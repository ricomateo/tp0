package communication

import "encoding/binary"

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

func serializeFinalizationMsg(agencyId string) []byte {
	messageType := byte(FinalizationMsg)
	serializedMsg := make([]byte, 0)
	serializedMsg = append(serializedMsg, messageType)

	// Serialize the batch size
	serializedMsg = serializeField(serializedMsg, agencyId)

	return serializedMsg
}

// serializeField serializes a string field and appends it to the given buffer
func serializeField(buf []byte, field string) []byte {
	fieldLength := byte(len(field))
	buf = append(buf, fieldLength)
	buf = append(buf, []byte(field)...)
	return buf
}
