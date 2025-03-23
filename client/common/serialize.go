package common

import (
	"encoding/binary"
	"reflect"
)

func (b *BetInfo) serialize() []byte {
	messageType := byte(0)
	serializedMsg := make([]byte, 0)
	serializedMsg = append(serializedMsg, messageType)

	// Serialize the name
	nameLength := byte(len(b.Name))
	serializedMsg = append(serializedMsg, nameLength)
	serializedMsg = append(serializedMsg, []byte(b.Name)...)

	// Serialize the last name
	lastNameLength := byte(len(b.LastName))
	serializedMsg = append(serializedMsg, lastNameLength)
	serializedMsg = append(serializedMsg, []byte(b.LastName)...)

	// Serialize the document
	documentBytes := make([]byte, 4)
	documentLength := reflect.TypeOf(b.Document).Size()
	// Write the document to a byte array
	binary.BigEndian.PutUint32(documentBytes, b.Document)
	serializedMsg = append(serializedMsg, byte(documentLength))
	serializedMsg = append(serializedMsg, documentBytes...)

	// Serialize the date of birth
	dateOfBirthLength := byte(len(b.DateOfBirth))
	serializedMsg = append(serializedMsg, dateOfBirthLength)
	serializedMsg = append(serializedMsg, []byte(b.DateOfBirth)...)

	// Serialize the bet number
	numberBytes := make([]byte, 4)
	numberLength := reflect.TypeOf(b.Number).Size()
	// Write the document to a byte array
	binary.BigEndian.PutUint32(numberBytes, b.Number)
	serializedMsg = append(serializedMsg, byte(numberLength))
	serializedMsg = append(serializedMsg, numberBytes...)

	return serializedMsg
}
