package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var ErrClientNotFound = errors.New("Client not found")

var upgrader = websocket.Upgrader{}

// RoomMetadata is a public view of the room's Metadata.
//
// It is mostly changed during the WAITING state while users are joining the room and choosing their subnet
type RoomMetadata struct {
	// The number of subnets in this room
	NumSubnets int `json:"num_subnets"`

	// Each subnet can have up to 255 players
	// 192.168.N.M -> Name
	Subnets map[int]map[int]Name `json:"subnets"`

	// An index for player name to IP address (reverse index of SubnetPlayers)
	IPAddresses map[Name]IP `json:"ip_addresses"`
}

type Challenge struct {
	// The challenge's destination IP address
	DestIP string `json:"destIP"`

	// The challenge's source IP address
	SourceIP string `json:"sourceIP"`

	// The challenge's question
	Question string `json:"question"`

	// The challenge's answer
	Answer string `json:"answer"`
}

type ChallengeResult struct {
	// If the question has been answered correctly
	Correct bool `json:"correct"`

	// The time the question was answered
	Created time.Time `json:"answered"`
}

type Room struct {
	// Locks the room
	sync.RWMutex

	code string

	// --- Public room data --- //

	// Metadata of the room (always available)
	Metadata RoomMetadata `json:"Metadata"`

	// State is the current state of the room
	State PublicState

	// --- Private room data --- //

	// Clients is a map from client session ID to client
	Clients map[string]*Client

	// Outstanding challenges
	Challenges map[Challenge]ChallengeResult

	// Q/A Tables
	QATables map[Name]QATable
}

func NewRoom(code string) *Room {
	numSubnets := 4
	subnets := make(map[int]map[int]Name)
	for i := 1; i <= numSubnets; i++ {
		subnets[i] = make(map[int]Name)
	}

	return &Room{
		code: code,
		Metadata: RoomMetadata{
			NumSubnets:  4,
			Subnets:     subnets,
			IPAddresses: map[Name]IP{},
		},
		Clients:    make(map[string]*Client),
		Challenges: make(map[Challenge]ChallengeResult),
		QATables:   make(map[Name]QATable),
	}
}

const alphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// NewClient creates a new client and adds it to the room
func (room *Room) NewClient() *Client {
	// Lock the room
	room.Lock()
	defer room.Unlock()

	// Generate a new session ID
	sessionID := make([]byte, 32)
	for i := range sessionID {
		sessionID[i] = alphabet[rand.Intn(len(alphabet))]
	}
	id := string(sessionID)

	// Generate a new name (that isn't already taken)
	var name Name
	for {
		name = NewName()

		// Check if the name is already taken
		taken := false
		for _, client := range room.Clients {
			if client.Name == name {
				taken = true
				break
			}
		}

		if !taken {
			break
		}
	}

	// Create the client
	room.Clients[id] = NewClient(id, name)

	// Create the Q/A table
	room.QATables[name] = NewQATable()

	// Return the session ID and name
	return room.Clients[id]
}

// AddConnection adds a websocket connection to the appropriate client
func (room *Room) AddConnection(sessionID string, conn *websocket.Conn) error {
	room.Lock()
	client, ok := room.Clients[sessionID]
	if !ok {
		room.Unlock()
		return ErrClientNotFound
	}
	client.AddConnection(conn)
	room.Unlock()
	return nil
}

type RoomUserData struct {
	// The user's name
	Name Name `json:"name"`

	// The user's IP address
	IP *IP `json:"ip,omitempty"`

	// The user's score
	Score int `json:"score"`

	// The user's Q/A table
	QATable QATable `json:"qa_table,omitempty"`
}

func (room *Room) UserData(client *Client) RoomUserData {
	room.RLock()
	defer room.RUnlock()

	// Get the user's IP address
	var result_ip *IP
	if ip, ok := room.Metadata.IPAddresses[client.Name]; ok {
		result_ip = &ip
	} else {
		result_ip = nil
	}

	// Get the user's score
	score := 0
	for challenge, result := range room.Challenges {
		if challenge.SourceIP == result_ip.String() && result.Correct {
			score++
		}
	}

	// Get the user's Q/A table
	qaTable, ok := room.QATables[client.Name]
	if !ok {
		qaTable = nil
	}

	return RoomUserData{
		Name:    client.Name,
		IP:      result_ip,
		Score:   score,
		QATable: qaTable,
	}
}
