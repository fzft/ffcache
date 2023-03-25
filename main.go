package main

import (
	"flag"
	"github.com/fzft/ffcache/cache"
	"log"
	"net"
	"time"
)

func main() {
	listenAddr := flag.String("listen", ":3000", "listen address")
	leaderAddr := flag.String("leader", "", "leader address")
	flag.Parse()

	opts := ServerOpts{
		ListenAddr: *listenAddr,
		IsLeader:   true,
		LeaderAddr: *leaderAddr,
	}

	go func() {
		time.Sleep(1 * time.Second)
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			log.Fatalf("Error dialing: %s", err)
		}

		conn.Write([]byte("hello"))
	}()

	s := NewServer(opts, cache.New())
	s.Start()

}
