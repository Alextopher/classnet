package classnet

import (
	"encoding/json"
	"fmt"
	"time"
)

type MessageType int

const (
	// Client -> Server
	JoinSubnet MessageType = iota
	RequestChallenge
	Answer
	RequestMetaData

	// Server -> Client
	AssignedIP
	CreateChallenge
	ChallengeResult

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
	"RequestChallenge",
	"Answer",
	"RequestMetaData",

	"AssignedIP",
	"CreateChallenge",
	"ChallengeResult",

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

// MarshalJSON marshals the enum as a quoted json string
func (mt MessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

// UnmarshalJSON unmarshal a quoted json string to the enum value
func (mt *MessageType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	// Find the string in the array and set the enum value
	for i, v := range messageTypeStrings {
		if v == s {
			*mt = MessageType(i)
			return nil
		}
	}

	return fmt.Errorf("invalid message type: %s", s)
}

// Message type for all client/server messages
type Message struct {
	// The type of message
	Type MessageType `json:"type"`

	// The message payload
	Payload interface{} `json:"payload"`
}

func (m Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// JoinSubnetMessage is sent by the client when joining a subnet
type JoinSubnetMessage struct {
	// The subnet
	Subnet int `json:"subnet"`
}

// RequestChallengeMessage is sent by the client to request a new challenge
type RequestChallengeMessage struct{}

// AnswerMessage is sent by the client to answer a challenge
type AnswerMessage struct {
	// The destination IP address
	Destination string `json:"destination"`
	// The question being answered
	Question string `json:"question"`
	// The answer
	Answer string `json:"answer"`
}

// RequestMetaData is sent by the client to the server, asking for updated metadata
type RequestMetaDataChallenge struct{}

// ---- Server -> Client ---- //

// ChallengeMessage is sent by the server to provide a new challenge
type ChallengeMessage struct {
	// The destination IP address
	Destination string `json:"destination"`
	// The question
	Question string `json:"question"`
}

// AssignedIPMessage is sent by the server to confirm joining a subnet, and to assign an IP address
type AssignedIPMessage struct {
	// The IP address assigned to this client
	IP IP `json:"ip"`
}

func NewAssignedIPMessage(ip IP) Message {
	return Message{
		Type: AssignedIP,
		Payload: AssignedIPMessage{
			IP: ip,
		},
	}
}

// AnswerResult is sent by the server to confirm the answer to a challenge
type AnswerResult struct {
	// The destination IP address
	Destination string `json:"destination"`
	// The question being answered
	Question string `json:"question"`
	// If the answer was correct
	Correct bool `json:"correct"`
}

func NewAnswerResultMessage(dest, question string, correct bool) Message {
	return Message{
		Type: ChallengeResult,
		Payload: AnswerResult{
			Destination: dest,
			Question:    question,
			Correct:     correct,
		},
	}
}

// StartMessage is sent by the host indicating when the game is starting
type StartMessage struct {
	// The unix time (seconds) when the game will start
	StartTime time.Time `json:"start_time"`
}

// StopMessage is sent by the host to immediately stop the game (with 1 minute grace period)
type StopMessage struct {
	// The unix time (seconds) when the grace period will end
	StopTime time.Time `json:"stop_time"`
}

// RestartMessage is sent by the host. This evicts all clients from their subnets, prompting them to rejoin and get a new IP address
type RestartMessage struct{}

// DestroyMessage is sent by the host. This evicts all clients from the room, and destroys the room
type DestroyMessage struct{}

// ErrorMessage is sent by anyone to indicate an error
type ErrorMessage struct {
	// The error message
	Message string `json:"message"`
}

// NewError creates a new error message
func NewError(msg string) Message {
	return Message{
		Type: Error,
		Payload: ErrorMessage{
			Message: msg,
		},
	}
}
