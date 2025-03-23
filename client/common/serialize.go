package common

func (m *Message) serialize() []byte {
	switch m.messageType {
	case StoreBet:
		payload := m.payload.(BetInfo)
		return payload.serialize()
	}
	return []byte{}
}

func (b *BetInfo) serialize() []byte {
	messageType := byte(0)
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

func serializeField(buf []byte, field string) []byte {
	fieldLength := byte(len(field))
	buf = append(buf, fieldLength)
	buf = append(buf, []byte(field)...)
	return buf
}
