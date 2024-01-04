package main

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrUnexpectedType = errors.New("unexpected message type")

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

func (m *Message) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Type    MessageType     `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.Type = aux.Type
	switch m.Type {
	case JoinSubnet:
		var payload JoinSubnetMessage
		if err := json.Unmarshal(aux.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case WhoAmI:
		var payload WhoAmIMessage
		if err := json.Unmarshal(aux.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case RequestChallenge:
		var payload RequestChallengeMessage
		if err := json.Unmarshal(aux.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case Answer:
		var payload AnswerMessage
		if err := json.Unmarshal(aux.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case RequestMetadata:
		var payload RequestMetadataMessage
		if err := json.Unmarshal(aux.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	case RequestUserdata:
		var payload RequestUserdataMessage
		if err := json.Unmarshal(aux.Payload, &payload); err != nil {
			return err
		}
		m.Payload = payload
	}

	return nil
}

// JoinSubnetMessage is sent by the client when joining a subnet
type JoinSubnetMessage struct {
	// The subnet
	Subnet int `json:"subnet"`
}

// WhoAmIMessage is sent by the client to request their own IP address
type WhoAmIMessage struct{}

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

// RequestMetaData is sent by the client to the server, asking for updated Metadata
type RequestMetadataMessage struct{}

// RequestUserdata is sent by the client to the server, asking for updated user data
type RequestUserdataMessage struct{}

// ---- Server -> Client ---- //

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

// CreateChallengeMessage is sent by the server to provide a new challenge
type CreateChallengeMessage struct {
	// The destination IP address
	Destination string `json:"destination"`
	// The question
	Question string `json:"question"`
}

func NewCreateChallengeMessage(dest, question string) Message {
	return Message{
		Type: CreateChallenge,
		Payload: CreateChallengeMessage{
			Destination: dest,
			Question:    question,
		},
	}
}

// Grade is sent by the server to confirm the answer to a challenge
type GradeMessage struct {
	// The destination IP address
	Destination string `json:"destination"`
	// The question being answered
	Question string `json:"question"`
	// If the answer was correct
	Correct bool `json:"correct"`
}

func NewGradeMessage(dest, question string, correct bool) Message {
	return Message{
		Type: Grade,
		Payload: GradeMessage{
			Destination: dest,
			Question:    question,
			Correct:     correct,
		},
	}
}

// MetadataMessage is sent by the server to provide complete and up-to-date Metadata
func NewMetadataMessage(metadata RoomMetadata) Message {
	return Message{
		Type:    Metadata,
		Payload: metadata,
	}
}

// UserDataMessage is sent by the server to provide complete and up-to-date user data
func NewUserdataMessage(userdata RoomUserData) Message {
	return Message{
		Type:    Userdata,
		Payload: userdata,
	}
}

// GameStateMessage is sent by the server to provide the current game state
// func NewGameStateMessage(state RoomState) Message {}

// ---- Host -> All ---- //

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
