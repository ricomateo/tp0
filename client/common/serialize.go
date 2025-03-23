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
	// Serialize the agency
	agencyLength := byte(len(b.Agency))
	serializedMsg = append(serializedMsg, agencyLength)
	serializedMsg = append(serializedMsg, []byte(b.Agency)...)

	// Serialize the name
	nameLength := byte(len(b.Name))
	serializedMsg = append(serializedMsg, nameLength)
	serializedMsg = append(serializedMsg, []byte(b.Name)...)

	// Serialize the last name
	lastNameLength := byte(len(b.LastName))
	serializedMsg = append(serializedMsg, lastNameLength)
	serializedMsg = append(serializedMsg, []byte(b.LastName)...)

	// Serialize the document
	documentLength := byte(len(b.Document))
	serializedMsg = append(serializedMsg, documentLength)
	serializedMsg = append(serializedMsg, []byte(b.Document)...)

	// Serialize the date of birth
	dateOfBirthLength := byte(len(b.DateOfBirth))
	serializedMsg = append(serializedMsg, dateOfBirthLength)
	serializedMsg = append(serializedMsg, []byte(b.DateOfBirth)...)

	// Serialize the bet number
	numberLength := byte(len(b.Number))
	serializedMsg = append(serializedMsg, numberLength)
	serializedMsg = append(serializedMsg, []byte(b.Number)...)

	return serializedMsg
}
