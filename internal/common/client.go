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

func NewClient(done chan struct{}, poolAddr string, poolPort int) (client *Client) {
	// TODO: need to formalize done channels throughout
	client = &Client{
		done:       done,
		poolAddr:   net.JoinHostPort(poolAddr, strconv.Itoa(poolPort)),
		connected:  false,
		joined:     false,
		sendStream: make(chan string, 0),
		broker:     NewBroker(done),
		mu:         new(sync.Mutex),
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go client.recv(done, &wg)
	go client.send(done, &wg)
	wg.Wait()

	return client
}

type Client struct {
	// TODO: Evaluate using a context instead of done channel
	done       chan struct{}
	poolAddr   string
	conn       net.Conn
	connected  bool
	joined     bool
	sendStream chan string
	broker     *Broker
	mu         *sync.Mutex
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
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

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
		select {
		case <-c.done:
		case c.sendStream <- msg:
		}

	}(msg)
}

func (c *Client) send(done chan struct{}, wg *sync.WaitGroup) {
	wg.Done()
	for {
		select {
		case <-done:
			return
		case msg := <-c.sendStream:
			// fmt.Println("Pulled from sendStream: ", msg)
			fmt.Fprint(c.conn, msg+"\n")
		}
	}
}

func (c *Client) recv(done chan struct{}, wg *sync.WaitGroup) {
	wg.Done()

	// TODO: handle this with sync.Cond or similar
loop:
	for {
		select {
		case <-done:
			return
		case <-time.After(100 * time.Millisecond):
			c.mu.Lock()
			if c.connected {
				break loop
			}
			c.mu.Unlock()
		}
	}

	scanner := bufio.NewScanner(c.conn)

	for scanner.Scan() {
		resp := scanner.Text()
		// fmt.Println("Got this from the svr: ", resp)
		msg, err := parse(resp)
		if err != nil {
			// fmt.Println("Received an unknown response: ", resp)
		}
		// fmt.Println("Parsed msg: ", msg)
		c.broker.Publish(msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "err reading scanner: ", err)
	}
}
