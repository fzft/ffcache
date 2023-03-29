package raft

type Transport interface {
	Connect(Transport)

	// ID returns the id of the transport
	ID() string
}
