package statler

import (
	"bufio"
	"bytes"
	"net"

	"github.com/ugorji/go/codec"
)

// Kind encodes the kind of statistic.
type Kind uint8

const (
	// Count indicates that statistic increments a counter value.
	Count Kind = iota
	// Value indicates that statistic sets a value.
	Value
)

// Stat holds a measurement to be forwarder to the server.
type Stat struct {
	Kind  Kind
	Count int32
	Value float64
	Key   string
}

// Client struct is used for sending statistics to the specified server.
type Client struct {
	c *net.UDPConn
	h codec.Handle
}

// NewClient creates and configures a new client.
func NewClient(address string) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	var h codec.Handle = new(codec.MsgpackHandle)

	return &Client{c: c, h: h}, nil
}

func (s Client) sendStat(stat *Stat) error {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)

	enc := codec.NewEncoder(bw, s.h)
	if err := enc.Encode(stat); err != nil {
		return err
	}

	bw.Flush()

	if _, err := s.c.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

// Value sends a value for the specified key.
func (s Client) Value(key string, value float64) error {
	return s.sendStat(&Stat{Kind: Value, Value: value, Key: key})
}

// Count increments a counter key by value.
func (s Client) Count(key string, count int32) error {
	return s.sendStat(&Stat{Kind: Count, Count: count, Key: key})
}

// Increment increments a counter key by one.
func (s Client) Increment(key string) error {
	return s.Count(key, 1)
}
