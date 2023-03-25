package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fzft/ffcache/cache"
	"io"
	"time"
)

type CommandType byte

const (
	TypeCMDSet CommandType = iota
	TypeCMDGet
	TypeCMDDel
)

type IExecuteCommand interface {
	Execute(cache.Cacher) *CMDStatus
	Bytes() []byte
	Parse(reader io.Reader) error
}

type BaseCommand struct {
	Cmd    CommandType
	Key    []byte
	rawMsg []byte

	Err error
}

func NewBaseCMD() *BaseCommand {
	return &BaseCommand{}
}

// tokenizer
// 1. split by space
// 2. check the first part is a valid command
// 3. check the second part is a valid key
// 4. check the third part is a valid value
// 5. check the fourth part is a valid ttl

// ParseRoute parses the raw message and returns the corresponding command
func (m *BaseCommand) ParseRoute(r io.Reader) (IExecuteCommand, error) {
	var t CommandType
	if err := binary.Read(r, binary.LittleEndian, &t); err != nil {
		return nil, err
	}

	var cmd IExecuteCommand
	switch t {
	case TypeCMDSet:
		cmd = newCMDSet()
	case TypeCMDGet:
		cmd = newCMDGet()
	default:
		return nil, fmt.Errorf("invalid command type: %d", t)
	}
	err := cmd.Parse(r)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (m *BaseCommand) Buffer() (*bytes.Buffer, error) {
	var err error
	buf := new(bytes.Buffer)
	// write command type
	if err = binary.Write(buf, binary.LittleEndian, m.Cmd); err != nil {
		return nil, err
	}

	// write key length
	if err = binary.Write(buf, binary.LittleEndian, int32(len(m.Key))); err != nil {
		return nil, err
	}

	// write key
	if err = binary.Write(buf, binary.LittleEndian, m.Key); err != nil {
		return nil, err
	}

	return buf, nil
}

type CMDStatus struct {
	*BaseCommand

	// val is output of the command
	val []byte
}

func newCMDStatus(val []byte, err error) *CMDStatus {
	return &CMDStatus{
		val: val,
		BaseCommand: &BaseCommand{
			Err: err,
		},
	}
}

type CMDSet struct {
	*BaseCommand

	// Value is used for set command
	Value []byte
	TTL   time.Duration
}

func newCMDSet() *CMDSet {
	return &CMDSet{
		BaseCommand: &BaseCommand{
			Cmd: TypeCMDSet,
		},
	}
}

func (m *CMDSet) Parse(r io.Reader) error {
	var (
		keyLen int32
		valLen int32
		err    error
	)

	if err = binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
		return err
	}
	m.Key = make([]byte, keyLen)
	if err = binary.Read(r, binary.LittleEndian, &m.Key); err != nil {
		return err
	}
	if err = binary.Read(r, binary.LittleEndian, &valLen); err != nil {
		return err
	}
	m.Value = make([]byte, valLen)
	if err = binary.Read(r, binary.LittleEndian, &m.Value); err != nil {
		return err
	}
	if err = binary.Read(r, binary.LittleEndian, &m.TTL); err != nil {
		return err
	}
	return nil
}

func (m *CMDSet) Bytes() []byte {
	buf, err := m.Buffer()
	if err != nil {
		return []byte{}
	}

	// write value length
	if err = binary.Write(buf, binary.LittleEndian, int32(len(m.Value))); err != nil {
		return []byte{}
	}

	// write value
	buf.Write(m.Value)

	// write ttl
	if err = binary.Write(buf, binary.LittleEndian, m.TTL); err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

func (m *CMDSet) Execute(c cache.Cacher) (stats *CMDStatus) {
	if err := c.Set(m.Key, m.Value, m.TTL); err != nil {
		stats.Err = err
		return
	}
	stats.val = []byte("OK")
	return
}

type CMDGet struct {
	*BaseCommand
}

func newCMDGet() *CMDGet {
	return &CMDGet{
		BaseCommand: &BaseCommand{
			Cmd: TypeCMDGet,
		},
	}
}

// Parse parses the raw message
func (m *CMDGet) Parse(r io.Reader) error {
	var keyLen int32
	if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
		return err
	}

	m.Key = make([]byte, keyLen)
	if err := binary.Read(r, binary.LittleEndian, &m.Key); err != nil {
		return err
	}
	return nil
}

func (m *CMDGet) Bytes() []byte {
	buf, err := m.Buffer()
	if err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

// Execute executes the command
func (m *CMDGet) Execute(c cache.Cacher) (status *CMDStatus) {
	val, err := c.Get(m.Key)
	if err != nil {
		status.Err = err
		return
	}
	status.val = val
	return
}

// parseCommand parses the command from client
func parseMessage(r io.Reader) (IExecuteCommand, error) {
	msg := NewBaseCMD()
	return msg.ParseRoute(r)
}
