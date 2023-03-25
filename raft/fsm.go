package raft

type LogEntry struct {
}

// State is the state of the server
// TODO: implement this by FSM
type State struct {
	//currentTerm latest term server has seen
	currentTerm int

	// candidateId that received vote in current
	//term
	voteFor string

	//log entries; each entry contains command
	//for state machine, and term when entry
	logs []LogEntry

	// index of highest log entry known to be
	//committed
	commitIndex int
}
