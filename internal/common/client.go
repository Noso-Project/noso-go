package common

import (
	"net"
	"strconv"
)

func NewClient(poolAddr string, poolPort int) *Client {
	return &Client{
		poolAddr:  net.JoinHostPort(poolAddr, strconv.Itoa(poolPort)),
		connected: false,
	}
}

type Client struct {
	poolAddr  string
	conn      net.Conn
	connected bool
	joined    bool
}

func (c *Client) Connect() (err error) {

	c.conn, err = net.Dial("tcp", c.poolAddr)
	if err != nil {
		return err
	}

	// TODO: do this with sync.Cond broadcast maybe?
	c.connected = true

	err = c.join()
	if err != nil {
		return err
	}

	// TODO: do this with sync.Cond broadcast maybe?
	c.joined = true
	return nil
}

func (c *Client) join() error {
	// TODO: Need to use real values for vession and instanceId
	return c.Send("JOIN ng9.9.9 123456")
}

func (c *Client) Send(msg string) error {
	return nil
}
