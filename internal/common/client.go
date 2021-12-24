package common

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	JoinTimeout = 5 * time.Second
)

var (
	JoinTimeoutErr = errors.New("Timed out while attempting to join pool")
)

func NewClient(poolAddr string, poolPort int) (client *Client) {
	// TODO: need to formalize done channels throughout
	done := make(chan struct{}, 0)
	client = &Client{
		poolAddr:   net.JoinHostPort(poolAddr, strconv.Itoa(poolPort)),
		connected:  false,
		joined:     false,
		sendStream: make(chan string, 0),
		broker:     NewBroker(done),
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go client.recv(&wg)
	go client.send(&wg)
	wg.Wait()

	return client
}

type Client struct {
	poolAddr   string
	conn       net.Conn
	connected  bool
	joined     bool
	sendStream chan string
	broker     *Broker
}

func (c *Client) Connect() (err error) {

	joinStream := c.broker.Subscribe(JOINOK)
	// TODO: enable and use unsubscribe
	// defer c.broker.Unsubscribe(joinStream)

	c.conn, err = net.Dial("tcp", c.poolAddr)
	if err != nil {
		return err
	}

	// TODO: do this with sync.Cond broadcast maybe?
	c.connected = true

	c.join()

	select {
	case <-joinStream:
	case <-time.After(JoinTimeout):
		return JoinTimeoutErr
	}

	c.joined = true
	return nil
}

func (c *Client) join() {
	// TODO: Need to use real values for vession and instanceId
	c.Send("JOIN ng9.9.9 123456")
}

func (c *Client) Send(msg string) {
	go func(msg string) {
		// TODO: Should I make this timeout? Or use a done chan?
		c.sendStream <- msg
	}(msg)
}

func (c *Client) send(wg *sync.WaitGroup) {
	wg.Done()
	for {
		select {
		case msg := <-c.sendStream:
			fmt.Println("Pulled from sendStream: ", msg)
			fmt.Fprint(c.conn, msg+"\n")
		}
	}
}

func (c *Client) recv(wg *sync.WaitGroup) {
	wg.Done()

	// TODO: handle this with sync.Cond or similar
loop:
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			if c.connected {
				break loop
			}
		}
	}

	scanner := bufio.NewScanner(c.conn)

	for scanner.Scan() {
		resp := scanner.Text()
		fmt.Println("Got this from the svr: ", resp)
		msg, err := parse(resp)
		if err != nil {
			fmt.Println("Received an unknown response: ", resp)
		}
		fmt.Println("Parsed msg: ", msg)
		c.broker.Publish(msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "err reading scanner: ", err)
	}
}
