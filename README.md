# CLASSNET

## Server progress

- [ ] Workflow to creates a new room
- [ ] Determines the number of "subnets"
- [ ] Chooses the number of messages required to be sent/received (the goal)
- [ ] Stop/start/reset the game
- [ ] Workflow to destroy the room
- [x] Choose which "subnet" to join
- [x] Students are assigned an IP address within the subnet
- [x] Can prompt the server for a "challenge". The challenge is a destination IP address and a question.

### Hosting progress

- [ ] Visit /room/:id to get redirected and join the room
- [ ] Get given a generated question-answer table
  - [ ] For now the table is a random map from 4- [ ]digit hex strings to 4- [ ]digit hex strings
  - [ ] I like the idea of using single symbols instead of strings
- [ ] By hand, they craft and send their question packet to the destination IP address.
- [ ] The expectation is another student will receive the packet, and will lookup the answer in their table.
- [ ] This second student will then send the answer back to the first student.
- [ ] Once the first student receives the answer to their question, they upload it to the website where it is verified and counted.
- [ ] While waiting for a response, students are expected to create more challenges and send more packets to other students.
- [ ] When time runs out, no more challenges can be created, but submissions will continue to be accepted for 1 more minute.

The goal is for the class to answer questions as quickly as possible.

### Definitions

Subnet: Within this project, I'm using 192.168.N.0/24 as the subnets. The first 2 bytes are fixed, the third byte is the subnet number, and the last byte is the host number.
