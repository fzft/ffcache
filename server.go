package main

import (
	"context"
	"github.com/fzft/ffcache/cache"
	"log"
	"net"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts

	followers map[net.Conn]struct{}

	cache cache.Cacher
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		cache:      c,
		followers:  make(map[net.Conn]struct{}),
	}
}

// Start listen on tcp port and start to serve
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}

	log.Printf("Server is listening on %s", s.ListenAddr)

	if !s.IsLeader {
		go func() {
			conn, err := net.Dial("tcp", s.LeaderAddr)
			if err != nil {
				log.Fatalf("Error dialing: %s", err)
			}
			s.handleConn(conn)
		}()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	if s.IsLeader {
		s.followers[conn] = struct{}{}
		defer func() {
			delete(s.followers, conn)
		}()
	}

	for {
		s.handleCommand(conn)
	}
}

// handleCommand handles the command from client
func (s *Server) handleCommand(conn net.Conn) error {

	var resp []byte

	cmd, err := parseMessage(conn)
	if err != nil {
		log.Printf("Error parsing command: %s", err)
		resp = []byte(err.Error())
		conn.Write(resp)
		return err
	}

	// execute the command , suppose it is a long time operation
	go func() {
		ack := cmd.Execute(s.cache)
		if ack.Err != nil {
			log.Printf("Error execute command: %s", err)
			resp = []byte(ack.Err.Error())
			conn.Write(resp)
			return
		}
		resp = ack.val
		conn.Write(resp)
	}()

	// send the command to followers
	go s.sendToFollowers(context.Background(), cmd)

	return nil
}

// sendToLeader sends the message to the leader
func (s *Server) sendToFollowers(ctx context.Context, cmd IExecuteCommand) error {
	for conn := range s.followers {
		go func(con net.Conn) {
			_, err := con.Write(cmd.Bytes())
			if err != nil {
				log.Printf("Error writing to follower: %s\n", err)
				return
			}
		}(conn)

	}

	return nil
}
