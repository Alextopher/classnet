package main

import "time"

// RoomState tracks the life-cycle of a room
// The host controls when the room transitions between states by sending messages
//
// Waiting
// Starting
// Running
// Stopping
// Stopped
type RoomState int

const (
	// Waiting is when the host is waiting for players to join (only Metadata is available)
	Waiting RoomState = iota
	// Starting is when the host has started the game
	Starting
	// Running is when the game is running
	Running
	// Stopping is when the game is coming to an end
	Stopping
	// Stopped is when the game has ended
	Stopped
)

type RunningState struct {
	// Scores of each player, allowing for creating a leaderboard
	Scores map[Name]int `json:"scores"`

	// Total number of messages sent/received by all players
	Progress int `json:"progress"`
}

// PublicState is separate from Metadata because it is updated more frequently
// and is used to track the life-cycle of a single game.
//
// Values become available as the life-cycle progresses.
// All previous values remain available until the game resets.
type PublicState struct {
	// The current state of the room
	State RoomState `json:"state"`

	// StartTime is the time when the game starts/started
	//
	// 10 seconds from the time the host sends the Start message
	StartTime time.Time `json:"startTime,omitempty"`

	// Scoreboard is the current scoreboard
	//
	// Becomes available once the game starts
	Scoreboard map[Name]int `json:"scoreboard,omitempty"`

	// Progress is the total number of messages sent/received by all players (optional)
	//
	// Becomes available once the game starts
	Progress int `json:"progress,omitempty"`

	// Goal is the number of messages required to be sent/received (optional)
	Goal int `json:"goal,omitempty"`

	// EndTime is the time when the game will end (optional)
	EndTime time.Time `json:"endTime,omitempty"`
}
