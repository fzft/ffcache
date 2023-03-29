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
	Id    string
	state *State

	lock sync.Mutex
	// peers is a map of server id to transport
	peers     map[string]*Server
	role      Role
	consumeCh chan any

	stopHeartBeatCh chan struct{}
	// resetCh is used to reset election timer
	resetCh chan struct{}

	//stopCh chan struct{}
	stopCh chan struct{}

	// counter used for vote
	counter int
}

func NewServer() *Server {
	s := &Server{
		Id:              uuid.New().String(),
		state:           &State{},
		role:            Follower,
		peers:           make(map[string]*Server),
		consumeCh:       make(chan any, 1024),
		stopHeartBeatCh: make(chan struct{}),
		resetCh:         make(chan struct{}),
		stopCh:          make(chan struct{}),
	}

	return s
}

// Start starts the server
func (s *Server) Start() {
	ticker := time.NewTicker(electionTimeout)
	defer ticker.Stop()
	// start election timer
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.electionTimeout()
		case <-s.resetCh:
			ticker.Reset(electionTimeout)
		}
	}
}

func (s *Server) electionTimeout() {
	// election timeout, start election
	s.lock.Lock()
	if s.role == Follower {
		s.role = Candidate
	}

	// increment currentTerm
	s.state.currentTerm += 1
	// vote for self
	s.state.voteFor = s.Id
	s.lock.Unlock()
	s.RequestVote()
}

// RequestVote requests votes from other servers,
// this method will register itself as a RPC handler
func (s *Server) RequestVote() {
	// broadcast requestVote RPC to all other servers
	for _, peer := range s.peers {
		peer.consumeCh <- RequestVote{
			term:         s.state.currentTerm,
			candidateId:  s.Id,
			lastLogIndex: s.state.preLogIndex,
			lastLogTerm:  s.state.preLogTerm,
		}
	}
	return
}

// AppendEntries is invoked by leader to replicate log entries,
// this method will register itself as a RPC handler
func (s *Server) AppendEntries() {
	// broadcast appendEntries RPC to all other servers
	for _, peer := range s.peers {
		peer.consumeCh <- AppendEntries{
			term:         s.state.currentTerm,
			leaderId:     s.Id,
			leaderCommit: s.state.commitIndex,
			prevLogTerm:  s.state.preLogIndex,
			prevLogIndex: s.state.preLogTerm,
		}
	}
	return
}

// Consume returns the channel to consume RPC
func (s *Server) handleMessage() {
	for {
		select {
		case msg := <-s.consumeCh:
			switch v := msg.(type) {
			case RequestVote:
				s.handleRequestVote(v)
			case AppendEntries:
				s.handleAppendEntries(v)
			case AppendEntriesResponse:
				s.handleAppendEntriesResponse(v)
			case VoteResponse:
				s.handleVoteResponse(v)
			}
		}
	}
}

// ID returns the id of the server
func (s *Server) ID() string {
	return s.Id
}

// Connect connects to other servers
func (s *Server) Connect(peer Transport) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.peers[peer.ID()] = peer.(*Server)
}

// heartbeat sends heartbeats to other servers, only called by leader
func (s *Server) heartbeat() {

	s.lock.Lock()
	if s.role != Leader {
		return
	}
	s.lock.Unlock()

	ticker := time.NewTicker(heartbeatTimeout)
	for {
		select {
		case <-s.stopHeartBeatCh:
			return
		case <-ticker.C:
			// broadcast appendEntries RPC with empty entries to all other servers
			s.AppendEntries()
		}
	}
}

func (s *Server) handleRequestVote(v RequestVote) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if v.term < s.state.currentTerm {
		s.peers[v.candidateId].consumeCh <- VoteResponse{
			term:        s.state.currentTerm,
			voteGranted: false,
		}
		return
	}

	// vote for self
	if s.state.voteFor == "" || v.candidateId == s.state.voteFor {
		//candidate's log is at least as up-to-date as receiver's log
		if v.lastLogIndex >= len(s.state.logs)-1 && v.lastLogTerm >= s.state.preLogTerm {
			s.state.voteFor = v.candidateId
			s.peers[v.candidateId].consumeCh <- VoteResponse{
				term:        s.state.currentTerm,
				voteGranted: true,
			}
			return
		}
	}
}

func (s *Server) handleAppendEntries(v AppendEntries) {
	s.lock.Lock()
	defer s.lock.Unlock()
	defer func() {
		<-s.resetCh
	}()

	if v.term < s.state.currentTerm || v.prevLogTerm != s.state.preLogTerm {
		s.peers[v.leaderId].consumeCh <- AppendEntriesResponse{
			term:    s.state.currentTerm,
			success: false,
		}
		return
	}

	// check self log entries has specified prevLogIndex
	if len(s.state.logs)-1 == v.prevLogIndex {
		//if self log index == prevLogIndex
		s.peers[v.leaderId].consumeCh <- AppendEntriesResponse{
			term:    s.state.currentTerm,
			success: false,
		}
		return
	} else {
		// update self log entries
	}

	// update commitIndex
	if v.leaderCommit > s.state.commitIndex {
		s.state.commitIndex = min(v.leaderCommit, len(s.state.logs)-1)

	}

	s.peers[v.leaderId].consumeCh <- AppendEntriesResponse{
		term:    s.state.currentTerm,
		success: true,
	}

	return
}

func (s *Server) handleAppendEntriesResponse(v AppendEntriesResponse) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if v.term > s.state.currentTerm {
		if s.role == Leader {
			s.stopHeartBeatCh <- struct{}{}
		}
		s.role = Follower
		return
	}

}

func (s *Server) handleVoteResponse(v VoteResponse) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if v.term == s.state.currentTerm {
		s.counter++
	}

	if s.counter > len(s.peers)/2 {
		if s.role != Leader {
			s.role = Leader
			go s.heartbeat()
		}
	}
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
