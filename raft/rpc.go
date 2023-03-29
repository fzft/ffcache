package raft

// RequestVote Invoked by candidates to gather votes
type RequestVote struct {

	// candidate’s term
	term int

	// candidate requesting vote
	candidateId string

	// index of candidate’s last log entry
	lastLogIndex int

	// term of candidate’s last log entry
	lastLogTerm int
}

type VoteResponse struct {
	//term for candidate to update itself
	term int

	// true means candidate received
	voteGranted bool
}

// AppendEntries Invoked by leader to replicate log entries
type AppendEntries struct {

	// leader’s term
	term int

	// so follower can redirect clients
	leaderId string

	prevLogIndex int

	prevLogTerm int

	leaderCommit int
}

type AppendEntriesResponse struct {

	// term for leader to update itself
	term int

	// true if follower contained entry matching
	success bool
}
