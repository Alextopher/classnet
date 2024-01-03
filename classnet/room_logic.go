package classnet

import "math/rand"

// Handles a client, expected to be run in a goroutine
func (room *Room) HandleClient(client *Client) {
	defer client.Close()

	// Send messages to the client
	go client.WritePump()

	// Receive messages from the client
	for msg := range client.receive {
		// Handle the message
		switch msg.Type {
		case JoinSubnet:
			room.JoinSubnet(client, msg.Payload.(JoinSubnetMessage))
		case RequestChallenge:
			room.RequestChallenge(client, msg.Payload.(RequestChallengeMessage))
		case Answer:
			room.Answer(client, msg.Payload.(AnswerMessage))
		}
	}
}

func (room *Room) JoinSubnet(client *Client, msg JoinSubnetMessage) {
	// Invalid subnet
	if msg.Subnet <= 0 || msg.Subnet > room.Metadata.Subnets {
		client.Send(NewError("Invalid subnet"))
	}

	room.Lock()
	// Remove the client from its existing subnet
	ip, ok := room.Metadata.IPAddresses[client.Name]
	if ok {
		delete(room.Metadata.SubnetPlayers[ip.Subnet], ip.Host)
	}

	// Choose the smallest host number that is not taken
	for host := 1; host <= 255; host++ {
		if _, ok := room.Metadata.SubnetPlayers[msg.Subnet][host]; !ok {
			// Found a free host number
			room.Metadata.SubnetPlayers[msg.Subnet][host] = client.Name
			room.Metadata.IPAddresses[client.Name] = IP{msg.Subnet, host}
			_ = client.Send(NewAssignedIPMessage(IP{msg.Subnet, host}))
			break
		}
	}
	room.Unlock()
}

func (room *Room) RequestChallenge(client *Client, msg RequestChallengeMessage) {
	// Get a random host that isn't the client
	room.Lock()
	sourceIP := room.Metadata.IPAddresses[client.Name]

	var destIP IP
	for {
		// Choose a random subnet
		destSubnet := rand.Intn(room.Metadata.Subnets) + 1
		subnet := room.Metadata.SubnetPlayers[destSubnet]

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
	room.Challenges[challenge] = struct{}{}
	room.Unlock()
}

// Answer receives a challenge answer from a client and verifies it
func (room *Room) Answer(client *Client, msg AnswerMessage) {
	room.Lock()

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
	if _, ok := room.Challenges[challenge]; !ok {
		// Challenge doesn't exist
		_ = client.Send(NewError("Challenge doesn't exist"))
		room.Unlock()
		return
	}

	// If the user guessed the right answer then we remove the challenge
	correct := msg.Answer == challenge.Answer
	if correct {
		delete(room.Challenges, challenge)
	}

	// Send the user a response, communicating if they got the answer right
	_ = client.Send(NewAnswerResultMessage(msg.Destination, msg.Question, correct))
}
