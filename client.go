package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client is an abstraction over any number of websocket connections that are tied to a single user
//
// Those connections maybe from different devices, but they all fan-in into a single Receive channels
// and fan-out from a single Send channel
type Client struct {
	// Locks the client's connections
	sync.Mutex

	// Websocket connections
	connections map[*websocket.Conn]struct{}

	// Outbound messages
	send chan []byte

	// Inbound messages being received from the client
	receive chan Message

	// A copy of the client's session ID
	SessionID string

	// A copy of the client's name
	Name Name
}

// NewClient creates a new client
func NewClient(sessionID string, name Name) *Client {
	return &Client{
		connections: make(map[*websocket.Conn]struct{}),
		receive:     make(chan Message),
		send:        make(chan []byte),
		SessionID:   sessionID,
		Name:        name,
	}
}

// AddConnection adds a new websocket connection to this client
func (c *Client) AddConnection(conn *websocket.Conn) {
	c.Lock()
	c.connections[conn] = struct{}{}
	c.Unlock()
	go c.ReadPump(conn)
}

// RemoveConnection removes a websocket connection from this client
func (c *Client) RemoveConnection(conn *websocket.Conn) {
	c.Lock()
	delete(c.connections, conn)
	c.Unlock()
}

// Send sends a message to the client
func (c *Client) Send(msg Message) error {
	log.Printf("Sending message to %s: %v\n", c.Name, msg)

	// JSON encode the message
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message: %v\n", err)
		return err
	}
	c.send <- bytes
	return nil
}

// SendBytes sends a byte array to the client
func (c *Client) SendBytes(bytes []byte) {
	log.Printf("Sending bytes to %s: %s\n", c.Name, bytes)

	c.send <- bytes
}

// WritePump pumps messages from the client's send channel to all of its connections
func (c *Client) WritePump() {
	log.Printf("Starting write pump for %s\n", c.Name)

	for bytes := range c.send {
		log.Printf("Sending bytes to the %d connections of %s\n", len(c.connections), c.Name)
		c.Lock()

		// Timeout 2 seconds
		deadline := time.Now().Add(2 * time.Second)

		var wg sync.WaitGroup
		for conn := range c.connections {
			wg.Add(1)
			go func(conn *websocket.Conn) {
				conn.SetWriteDeadline(deadline)
				conn.WriteMessage(websocket.TextMessage, bytes)
				wg.Done()
			}(conn)
		}
		wg.Wait()
		c.Unlock()
	}
}

// ReadPump starts a goroutine to read messages from a connection and send them to the client's receive channel
func (c *Client) ReadPump(conn *websocket.Conn) {
	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		log.Printf("Received bytes from %s: %s\n", c.Name, bytes)

		// Decode the message
		var msg Message
		err = json.Unmarshal(bytes, &msg)
		if err != nil {
			// Send an error message to this connection
			log.Printf("Failed to decode message: %v\n", err)
			errBytes, _ := NewError("Failed to decode message").Marshal()
			conn.WriteMessage(websocket.TextMessage, errBytes)
			continue
		}

		log.Printf("Received message from %s: %v\n", c.Name, msg)

		// Forward the message to the client
		c.receive <- msg
	}
	c.RemoveConnection(conn)
	conn.Close()
}

// Close closes all of the client's connections
func (c *Client) Close() {
	c.Lock()
	for conn := range c.connections {
		conn.Close()
	}
	c.Unlock()
}
