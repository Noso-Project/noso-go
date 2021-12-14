package common

import (
	"errors"
	"net"
	"time"
)

const (
	connDialTimeout = 5 * time.Second
	connDeadline    = 20 * time.Second // Read/Write deadline
	// reconnectSleep    = 5 * time.Second
)

var (
	DIALTIMEOUTERR = errors.New("Connect() timed out")
)

type dialer func(string, string, time.Duration) (net.Conn, error)

func NewClient(addr string) *Client {
	return &Client{
		addr:            addr,
		connected:       false,
		connDialTimeout: connDialTimeout,
	}
}

type Client struct {
	addr            string
	connected       bool
	connDialTimeout time.Duration
	conn            net.Conn
}

func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.addr, c.connDialTimeout)
	if err != nil {
		return DIALTIMEOUTERR
	}

	c.conn = conn
	c.connected = true
	conn.SetDeadline(time.Now().Add(connDeadline))

	return nil
}

func (c *Client) Disconnect() {
	c.conn.Close()
	c.connected = false
}
