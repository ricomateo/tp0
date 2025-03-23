package common

type FieldType byte

const (
	NameOpcode FieldType = iota
	LastNameOpcode
	DocumentOpcode
	DateOfBirthOpcode
	NumberOpcode
)

func (b *BetInfo) serialize() []byte {
	serializedMsg := make([]byte, 0)
	serializedMsg = append(serializedMsg, byte(NameOpcode))
	serializedMsg = append(serializedMsg, byte(len(b.Name)))
	serializedMsg = append(serializedMsg, []byte(b.Name)...)
	// TODO: serialize the remaining fields
	return serializedMsg
}
