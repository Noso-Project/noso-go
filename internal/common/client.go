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

var (
	ConnectTimeout = 5 * time.Second
	JoinTimeout    = 5 * time.Second
	PingInterval   = 5 * time.Second
)

var (
	JoinTimeoutErr = errors.New("Timed out while attempting to join pool")
	PassFailedErr  = errors.New("Failed to join pool: wrong password")
)

func NewClient(done chan struct{}, poolAddr string, poolPort int) (client *Client) {
	// TODO: need to formalize done channels throughout
	client = &Client{
		done:         done,
		poolAddr:     net.JoinHostPort(poolAddr, strconv.Itoa(poolPort)),
		connected:    make(chan struct{}, 0),
		joined:       make(chan struct{}, 0),
		sendStream:   make(chan string, 0),
		broker:       NewBroker(done),
		mu:           new(sync.Mutex),
		joinTimeout:  JoinTimeout,
		pingInterval: PingInterval,
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go client.recv(done, &wg)
	go client.send(done, &wg)
	go client.ping(done, &wg)
	wg.Wait()

	return client
}

type Client struct {
	// TODO: Evaluate using a context instead of done channel
	done       chan struct{}
	poolAddr   string
	conn       net.Conn
	connected  chan struct{}
	joined     chan struct{}
	sendStream chan string
	broker     *Broker
	mu         *sync.Mutex

	// Timeouts/Intervals
	joinTimeout  time.Duration
	pingInterval time.Duration
}

func (c *Client) Connect() (err error) {

	joinStream := c.broker.Subscribe(JoinTopic)
	defer c.broker.Unsubscribe(joinStream)

	c.conn, err = net.DialTimeout("tcp", c.poolAddr, ConnectTimeout)
	if err != nil {
		return err
	}

	close(c.connected)

	c.join()

	select {
	case <-joinStream:
	case <-time.After(c.joinTimeout):
		return JoinTimeoutErr
	}

	close(c.joined)
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
			fmt.Fprintln(c.conn, msg)
		}
	}
}

func (c *Client) recv(done chan struct{}, wg *sync.WaitGroup) {
	wg.Done()

	select {
	case <-c.connected:
	case <-done:
		return
	}

	scanner := bufio.NewScanner(c.conn)

	for scanner.Scan() {
		resp := scanner.Text()
		// fmt.Println("Got this from the svr: ", resp)
		msg, err := parse(resp)
		if err != nil {
			fmt.Println("Received an unknown response: ", resp)
		}
		// fmt.Println("Parsed msg: ", msg)
		c.broker.Publish(msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "err reading scanner: ", err)
	}
}

func (c *Client) ping(done chan struct{}, wg *sync.WaitGroup) {
	joinStream := c.broker.Subscribe(JoinTopic)
	wg.Done()
	select {
	case <-done:
		return
	case <-joinStream:
		close(joinStream)
	}

	ticker := time.NewTicker(c.pingInterval)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			fmt.Fprintln(c.conn, "PING")
		}
	}
}
