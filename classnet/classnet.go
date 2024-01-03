package classnet

// CLASSNET

// Rooms have a host and a set of clients.
// Host:
// - Creates the room
// - Determines the number of "subnets"
// - Chooses the number of messages required to be sent/received (the goal)
// - Can stop/start/reset the game
// - Can destroy the room

// Students:
// - Enter a room-code
// - Choose which "subnet" to join
//    - Within a subnet (local area network), students can communicate _however they want_.
//    - Between subnets, students must use are CLASSNET ip-like packets
// - Students are assigned a random IP address within the subnet
// - Given a generated question-answer table
//   - For now the table is a random map from 4-digit hex strings to 4-digit hex strings
//   - I like the idea of using single symbols instead of strings
// - Can prompt the server for a "challenge". The challenge is a destination IP address and a question.
// - By hand, they craft and send their question packet to the destination IP address.
// - The expectation is another student will receive the packet, and will lookup the answer in their table.
// - This second student will then send the answer back to the first student.
// - Once the first student receives the answer to their question, they upload it to the website where it is verified and counted.

// - While waiting for a response, students are expected to create more challenges and send more packets to other students.
// - When time runs out, no more challenges can be created, but submissions will continue to be accepted for 1 more minute.

// The goal is for the class to answer questions as quickly as possible.
