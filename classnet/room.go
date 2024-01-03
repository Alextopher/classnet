package classnet

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var ErrClientNotFound = errors.New("Client not found")

var upgrader = websocket.Upgrader{}

type RoomState int

const (
	// Waiting is when the host is waiting for players to join (only metadata is available)
	Waiting RoomState = iota
	// Starting is when the host has started the game
	Starting
	// Running is when the game is running
	Running
	// Stopping is when the game is coming to an end
	Stopping
)

type IP struct {
	// Subnet
	Subnet int `json:"subnet"`
	// Host
	Host int `json:"host"`
}

func (ip IP) String() string {
	return fmt.Sprintf("192.168.%d.%d", ip.Subnet, ip.Host)
}

type Metadata struct {
	// The number of subnets in this room
	Subnets int `json:"num_subnets"`

	// Each subnet can have up to 255 players
	// 192.168.N.M -> Name
	SubnetPlayers map[int]map[int]Name `json:"subnets"`

	// An index for player name to IP address
	IPAddresses map[Name]IP `json:"ip_addresses"`

	// The number of messages required to be sent/received (the goal)
	Goal int `json:"goal"`
}

type WaitingState struct{}

type StartingState struct {
	// The time when the game will start
	StartTime time.Time `json:"startTime"`
}

type RunningState struct {
	// Scores of each player, allows us to create a leaderboard
	Scores map[string]int `json:"scores"`

	// Total number of messages sent/received
	Progress int `json:"progress"`
}

type StoppingState struct {
	// The time when the game will end
	EndTime time.Time `json:"endTime"`
}

type PublicState struct {
	// Type is type of the current state of the room
	Type RoomState `json:"type"`
	// State is the current state of the room
	State interface{} `json:"state"`
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

type Room struct {
	// Locks the room
	sync.RWMutex

	// --- Public room data --- //

	// Metadata of the room (always available)
	Metadata Metadata `json:"metadata"`

	// State is the current state of the room
	State PublicState

	// --- Private room data --- //

	// Clients is a map from client session ID to client
	Clients map[string]*Client

	// Outstanding challenges
	Challenges map[Challenge]struct{}

	// Q/A Tables
	QATables map[Name]QATable
}

const alphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// NewClient creates a new client and adds it to the room
func (room *Room) NewClient() (string, Name) {
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

	// Return the session ID and name
	return id, name
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

type Rooms struct {
	// Locks the rooms
	sync.RWMutex

	// Rooms is a map from room code to room
	Rooms map[string]*Room
}

// WebsocketHandler handles incoming websocket connections
// /room/{code}/ws
func WebsocketFactory(rooms *Rooms) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		code := vars["code"]
		if code == "" {
			http.Error(w, "Missing room code", http.StatusBadRequest)
			return
		}

		// Get the room object
		rooms.RLock()
		defer rooms.RUnlock()

		room, ok := rooms.Rooms[code]
		if !ok {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		// Get the client's session ID from the cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Missing session cookie", http.StatusBadRequest)
			return
		}

		// Upgrade the connection to a websocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection to websocket", http.StatusInternalServerError)
			return
		}

		// Add the connection to the client
		err = room.AddConnection(cookie.Value, conn)
		if err != nil {
			// Tell the client to ditch their cookie
			cookie := &http.Cookie{
				Name:    "session",
				Value:   "",
				Expires: time.Now().Add(-1 * time.Hour),
			}
			http.SetCookie(w, cookie)
			http.Error(w, "Invalid session cookie", http.StatusBadRequest)

			// Close the connection
			conn.Close()
			return
		}
	}
}

// RegisterHandler handles incoming client registrations
// /room/{code}/register
func RegisterFactory(rooms *Rooms) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		code := vars["code"]
		if code == "" {
			http.Error(w, "Missing room code", http.StatusBadRequest)
			return
		}

		rooms.RLock()
		defer rooms.RUnlock()

		// Get the room object
		room, ok := rooms.Rooms[code]
		if !ok {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		// Create a new client
		sessionID, _ := room.NewClient()

		// Set the session cookie
		cookie := &http.Cookie{
			Name:  "session",
			Value: sessionID,
			// Expires in 1 hour
			Expires: time.Now().Add(1 * time.Hour),
		}

		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusOK)
	}
}
