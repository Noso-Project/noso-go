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
		auth:         "password leviable5",
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
	// TODO: Set auth
	done       chan struct{}
	poolAddr   string
	auth       string // "poolPw walletAddr"
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

	// TODO: Might need to explicitely separate connect and join,
	//       as its possible a secondary client might want to connect,
	//       to a pool but not join it until the primary fails
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
	c.Send("JOIN ng9.9.9")
}

func (c *Client) Send(msg string) {
	go func(msg string) {
		// TODO: Should I make this timeout? Or use a context deadline?
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
			msg = c.auth + " " + msg
			fmt.Println("Send: ", msg)
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
		fmt.Println("Recv: ", resp)
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
			// TODO: Need to use real hash rate value here instead of zero
			c.Send("PING 0")
		}
	}
}
