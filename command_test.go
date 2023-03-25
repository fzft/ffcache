package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCMDGet_Parse(t *testing.T) {
	cmd := newCMDGet()
	cmd.Key = []byte("hello")

	buf := cmd.Bytes()
	r := bytes.NewReader(buf)

	//move to next byte
	r.ReadByte()

	cmd2 := newCMDGet()
	cmd2.Parse(r)
	assert.Equal(t, cmd2, cmd)
}

func TestCMDSet_Parse(t *testing.T) {
	cmd := newCMDSet()
	cmd.Key = []byte("hello")
	cmd.Value = []byte("world")
	cmd.TTL = 100

	buf := cmd.Bytes()
	r := bytes.NewReader(buf)

	//move to next byte
	r.ReadByte()

	cmd2 := newCMDSet()
	cmd2.Parse(r)

	assert.Equal(t, cmd2, cmd)
}
