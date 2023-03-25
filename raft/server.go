package raft

// raft implements the Raft consensus algorithm.
// https://raft.github.io/raft.pdf

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

const (
	electionTimeout  = 150 * time.Millisecond
	heartbeatTimeout = 50 * time.Millisecond
)

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type Server struct {
	id    string
	state *State

	lock sync.Mutex
	role Role
}

func NewServer() *Server {
	s := &Server{
		id:    uuid.New().String(),
		state: &State{},
		role:  Follower,
	}

	go s.Start()
	return s
}

// Start starts the server
func (s *Server) Start() {
	ticker := time.NewTicker(electionTimeout)
	defer ticker.Stop()
	// start election timer
	for {
		<-ticker.C
		s.lock.Lock()
		if s.role == Follower {
			s.role = Candidate
		}

		// increment currentTerm
		s.state.currentTerm += 1
		// vote for self
		s.state.voteFor = s.id
		s.lock.Unlock()
		s.RequestVote()
	}
}

// RequestVote requests votes from other servers,
//this method will register itself as a RPC handler
func (s *Server) RequestVote() {

	// broadcast requestVote RPC to all other servers
}

// AppendEntries is invoked by leader to replicate log entries,
//this method will register itself as a RPC handler
func (s *Server) AppendEntries() {

}

// heartbeat sends heartbeats to other servers, only called by leader
func (s *Server) heartbeat() {

	if s.role != Leader {
		return
	}

	ticker := time.NewTicker(heartbeatTimeout)
	for {
		<-ticker.C

		// broadcast appendEntries RPC with empty entries to all other servers
		s.AppendEntries()
	}
}
