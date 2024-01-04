package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Handles incoming messages from clients
func (room *Room) HandleClientMessages(client *Client) {
	// Receive messages from the client
	for msg := range client.receive {
		log.Printf("Handling message from %s: %v\n", client.Name, msg)

		// Handle the message
		switch msg.Type {
		case JoinSubnet:
			msg, ok := msg.Payload.(JoinSubnetMessage)
			if !ok {
				_ = client.Send(NewError("INVALID_PAYLOAD: Expected JoinSubnetMessage"))
				continue
			}
			room.JoinSubnet(client, msg)
		case WhoAmI:
			room.SendUserdata(client)
		case RequestChallenge:
			msg, ok := msg.Payload.(RequestChallengeMessage)
			if !ok {
				_ = client.Send(NewError("INVALID_PAYLOAD: Expected RequestChallengeMessage"))
				continue
			}
			room.RequestChallenge(client, msg)
		case Answer:
			msg, ok := msg.Payload.(AnswerMessage)
			if !ok {
				_ = client.Send(NewError("INVALID_PAYLOAD: Expected AnswerMessage"))
				continue
			}
			room.Answer(client, msg)
		case RequestMetadata:
			room.SendMetadata(client)
		}
	}

	client.Close()
}

// Broadcast sends a message to all clients in the room
//
// TODO: Checkout "prepared messages". Seems like it might be even better than
// caching the bytes.
func (room *Room) Broadcast(msg Message) {
	// Make the message some bytes
	msgBytes, _ := msg.Marshal()

	room.Lock()
	for _, client := range room.Clients {
		client.SendBytes(msgBytes)
	}
	room.Unlock()
}

// BroadcastMetadata sends the room Metadata to all clients in the room
func (room *Room) BroadcastMetadata() {
	room.Broadcast(NewMetadataMessage(room.Metadata))
}

// JoinSubnet is called to handle a JoinSubnet message
func (room *Room) JoinSubnet(client *Client, msg JoinSubnetMessage) {
	// Subnet joins are only allowed while the room is in "Waiting" state
	room.Lock()
	if room.State.State != Waiting {
		_ = client.Send(NewError(fmt.Sprintf("WRONG_STATE: Attempted to join subnet while game is not waiting (state: %d)", room.State.State)))
		room.Unlock()
		return
	}

	// Verify that the subnet is valid
	if msg.Subnet <= 0 || msg.Subnet > room.Metadata.NumSubnets {
		client.Send(NewError(fmt.Sprintf("INVALID_SUBNET: Subnet %d does not exist. Expected 1 <= subnet <= %d", msg.Subnet, room.Metadata.NumSubnets)))
		room.Unlock()
		return
	}

	// Remove the client from its existing subnet
	ip, ok := room.Metadata.IPAddresses[client.Name]
	if ok {
		delete(room.Metadata.Subnets[ip.Subnet], ip.Host)
	}

	// Choose the smallest host number that is not taken
	for host := 1; host <= 255; host++ {
		if _, ok := room.Metadata.Subnets[msg.Subnet][host]; !ok {
			// Found a free host number
			room.Metadata.Subnets[msg.Subnet][host] = client.Name
			room.Metadata.IPAddresses[client.Name] = IP{msg.Subnet, host}
			_ = client.Send(NewAssignedIPMessage(IP{msg.Subnet, host}))
			break
		}
	}
	room.Unlock()

	// This changes the room's Metadata, so it needs to be rebroadcasted
	room.SendUserdata(client)
	room.BroadcastMetadata()
}

// RequestChallenge is called to handle a RequestChallenge message
func (room *Room) RequestChallenge(client *Client, msg RequestChallengeMessage) {
	room.Lock()
	// Challenges can only be requested while the room is in "Running" state
	if room.State.State != Running {
		_ = client.Send(NewError(fmt.Sprintf("WRONG_STATE: Attempted to request challenge while game is not running (state: %d)", room.State.State)))
		room.Unlock()
		return
	}

	// Get a random host that isn't the client
	sourceIP := room.Metadata.IPAddresses[client.Name]
	var destIP IP
	for {
		// Choose a random subnet
		destSubnet := rand.Intn(room.Metadata.NumSubnets) + 1
		subnet := room.Metadata.Subnets[destSubnet]

		// Choose a random key from the subnet
		host, n := RandomEntry(subnet)

		// Verify that this destination is not the client
		if n != client.Name {
			destIP = IP{destSubnet, host}
			break
		}
	}

	// Generate a new challenge
	var challenge Challenge
	for {
		question, answer := RandomEntry(room.QATables[client.Name])
		challenge = Challenge{
			DestIP:   destIP.String(),
			SourceIP: sourceIP.String(),
			Question: question,
			Answer:   answer,
		}

		// Verify that this challenge doesn't already exist
		if _, ok := room.Challenges[challenge]; !ok {
			break
		}
	}

	// Add the challenge to the room
	room.Challenges[challenge] = ChallengeResult{
		Correct: false,
		Created: time.Now(),
	}

	// Send the challenge to the client
	_ = client.Send(NewCreateChallengeMessage(challenge.DestIP, challenge.Question))
	room.Unlock()
}

// Answer is called to handle an Answer message
//
// Communicates to the client if they got the answer right
func (room *Room) Answer(client *Client, msg AnswerMessage) {
	room.Lock()
	// Answers can only be accepting while the room is "Running" or "Stopping"
	if room.State.State != Running && room.State.State != Stopping {
		_ = client.Send(NewError(fmt.Sprintf("WRONG_STATE: Answers can only be accepted while the room is running or stopping (state: %d)", room.State.State)))
		room.Unlock()
		return
	}

	// Get the user's IP address
	ip := room.Metadata.IPAddresses[client.Name].String()

	// Check if the challenge exists
	challenge := Challenge{
		DestIP:   msg.Destination,
		SourceIP: ip,
		Question: msg.Question,
		Answer:   msg.Answer,
	}

	// Check if the challenge exists
	result, ok := room.Challenges[challenge]
	if !ok {
		// Challenge doesn't exist
		_ = client.Send(NewError("Challenge doesn't exist"))
		room.Unlock()
		return
	}

	// If the user guessed the right answer then we mark the challenge as solved
	correct := msg.Answer == challenge.Answer
	if correct {
		room.Challenges[challenge] = ChallengeResult{
			Correct: true,
			Created: result.Created,
		}
	}

	// Send the user a response, communicating if they got the answer right
	_ = client.Send(NewGradeMessage(msg.Destination, msg.Question, correct))
}

// SendMetadata sends the room Metadata to the client
func (room *Room) SendMetadata(client *Client) {
	// Send the client the Metadata
	_ = client.Send(NewMetadataMessage(room.Metadata))
}

// SendUserdata sends the room user data to the client
func (room *Room) SendUserdata(client *Client) {
	_ = client.Send(NewUserdataMessage(room.UserData(client)))
}
