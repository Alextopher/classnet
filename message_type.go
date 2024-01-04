package main

type MessageType int

const (
	// Client -> Server
	JoinSubnet MessageType = iota
	WhoAmI
	RequestChallenge
	Answer
	RequestMetadata
	RequestUserdata

	// Server -> Client
	AssignedIP
	CreateChallenge
	Grade
	Metadata
	Userdata

	// Host -> All
	Start
	Stop
	Restart
	Destroy

	// Generic Errors
	Error
)

var messageTypeStrings = [...]string{
	"JoinSubnet",
	"WhoAmI",
	"RequestChallenge",
	"Answer",
	"RequestMetadata",
	"RequestUserdata",

	"AssignedIP",
	"CreateChallenge",
	"Grade",
	"Metadata",
	"Userdata",

	"Start",
	"Stop",
	"Restart",
	"Destroy",

	"Error",
}

// String returns the string representation of the message type
func (mt MessageType) String() string {
	return messageTypeStrings[mt]
}

func (mt MessageType) MarshalText() ([]byte, error) {
	return []byte(mt.String()), nil
}

func (mt *MessageType) UnmarshalText(b []byte) error {
	s := string(b)
	for i, str := range messageTypeStrings {
		if s == str {
			*mt = MessageType(i)
			return nil
		}
	}
	return nil
}
