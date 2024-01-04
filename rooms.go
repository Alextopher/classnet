package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var rooms = &Rooms{
	Rooms: make(map[string]*Room),
}

type Rooms struct {
	// Locks the rooms
	sync.RWMutex

	// Rooms is a map from room code to room
	Rooms map[string]*Room
}

func (r *Rooms) NewRoom(code string) *Room {
	// If the code is empty, generate a random one
	if code == "" {
		code = randomSymbol()
	}
	room := NewRoom(code)

	r.Lock()
	r.Rooms[code] = room
	r.Unlock()

	return room
}

// WebsocketHandler handles incoming websocket connections
// /room/{code}/ws
func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
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
			Name:     "session",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		conn.Close()
		return
	}

	log.Println("Added connection to room", code)
}

// RegisterHandler handles incoming client registrations
// /room/{code}/register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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
	client := room.NewClient()
	go room.HandleClientMessages(client)
	go client.WritePump()

	// Set the session cookie
	cookie := &http.Cookie{
		Name:  "session",
		Value: client.SessionID,
		// Expires in 1 hour
		Expires: time.Now().Add(1 * time.Hour),
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}
