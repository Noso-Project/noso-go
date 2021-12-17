package common

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	// Timeout related
	connDialTimeout = 5 * time.Second
	joinOkTimeout   = 5 * time.Second
	connDeadline    = 20 * time.Second // Read/Write deadline
	sendDeadline    = 20 * time.Second // Read/Write deadline
	// reconnectSleep    = 5 * time.Second
)

const (
	// inter-client signaling messages
	JOINED clientMessage = iota + 1
	CONNECTED
)

var (
	DIALTIMEOUTERR    = errors.New("Connect() timed out")
	SENDTIMEOUTERR    = errors.New("Timed out writing to Send channel")
	JOINTIMEOUTERR    = errors.New("Timed out waiting for JOINOK response")
	supportedMessages = []clientMessage{JOINED, CONNECTED}
)

type clientMessage int

func NewClient(addr string) (client *Client) {
	client = &Client{
		addr: addr,
		// TODO: Need to pass these in and set them
		auth:            fmt.Sprintf("%s %s", "dummy-pool-pw", "dummy-wallet"),
		connDialTimeout: connDialTimeout,
		sendTimeout:     sendDeadline,
		joinOkTimeout:   joinOkTimeout,

		// Messaging
		Send:    make(chan string, 0),
		msgSubs: newClientMessageMap(),

		// State vars
		connected: false,
		joined:    false,
	}

	go client.recv()
	go client.send()

	return
}

type Client struct {
	addr            string
	auth            string
	connDialTimeout time.Duration
	sendTimeout     time.Duration
	joinOkTimeout   time.Duration
	conn            net.Conn

	// Messaging
	Send    chan string
	msgSubs clientMessageMap
	mu      sync.RWMutex

	// State vars
	connected bool
	joined    bool
}

func (c *Client) Connect() error {
	var err error

	if c.conn == nil {
		fmt.Println("Setting c.conn")
		c.conn, err = net.DialTimeout("tcp", c.addr, c.connDialTimeout)
		if err != nil {
			return DIALTIMEOUTERR
		}
	}

	// time.Sleep(500 * time.Millisecond)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("Connected, About to notify ")
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	c.connected = true
	c.conn.SetDeadline(time.Now().Add(connDeadline))
	c.Notify(CONNECTED)

	// TODO: figure out how to log without triggering test logs
	// log.Printf("Successfully connected to %s", c.addr)

	// TODO: Need to use real values for miner version and instanceId
	joined := c.Subscribe(JOINED)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("Sending JOIN ")
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	join := fmt.Sprintf("JOIN %s %s", "ng0.0.0", "123456")
	select {
	case c.Send <- join:
	case <-time.After(c.sendTimeout):
		return SENDTIMEOUTERR
	}

	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("Waiting for JOINOK")
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	select {
	case <-joined:
		c.joined = true
		return nil
	case <-time.After(c.joinOkTimeout):
		return JOINTIMEOUTERR
	}
}

func (c *Client) Disconnect() {
	c.conn.Close()
	c.connected = false
}

func (c *Client) Subscribe(msgType clientMessage) chan clientMessage {
	// TODO: Need to make a RWLock for this map
	// TODO: Will need to groom out dead channels?
	fmt.Println("-----------------------------------")
	fmt.Println("Subscribing to: ", msgType)
	fmt.Println("-----------------------------------")
	ch := make(chan clientMessage, 0)
	c.mu.Lock()
	c.msgSubs[msgType] = append(c.msgSubs[msgType], ch)
	c.mu.Unlock()
	return ch
}

func (c *Client) Notify(msgType clientMessage) {
	// TODO: Need to make a RWLock for this map
	// TODO: Will need to groom out dead channels?
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("About to notify: ", msgType)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	c.mu.Lock()
	notify := c.msgSubs[msgType][:]
	c.mu.Unlock()
	for _, ch := range notify {
		fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
		fmt.Printf("About to notify channel %v: %d\n", ch, msgType)
		fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
		go func(ch chan clientMessage) {
			select {
			case ch <- msgType:
			case <-ch:
				// Channel already closed
				// TODO: Groom out channels here?
				fmt.Println("&&&&&&&&&&&&&&&&&\nChannel already closed\n&&&&&&&&&&&&&&&&&")
			case <-time.After(1000 * time.Millisecond):
				// Channel still alive but other end isn't responding
				// TODO: Groom out channels here?
				// TODO: Close channel here?
				fmt.Println("*****************\nFailed to notify channel\n*****************")
			}
		}(ch)
	}
}

func (c *Client) send() {
	for {
		select {
		case msg := <-c.Send:
			msg = fmt.Sprintf("%s %s\n", c.auth, msg)
			fmt.Fprintf(c.conn, msg)
		}
	}
}

func (c *Client) recv() {
	// Wait for the connection to come up
	fmt.Println("WAITING TO CONNECT")
	<-c.Subscribe(CONNECTED)
	fmt.Println("CONNECTED")

	scanner := bufio.NewScanner(c.conn)

	for {
		if ok := scanner.Scan(); !ok {
			// TODO: handle this better
			fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
			fmt.Println("Error scanning connection: ", scanner.Err())
			fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
			break
		}
		msg := scanner.Text()
		c.parse(msg)
	}
}

func (c *Client) parse(raw string) {
	if raw == "" {
		// TODO: handle this better
		// panic("Got an empty message")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("Got an empty message")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		return
	}

	msg := strings.Split(raw, " ")

	switch msg[0] {
	case "JOINOK":
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println("Got a JOINOK")
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		c.joined = true
		c.Notify(JOINED)
	default:
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println("Not sure what this is: ", msg)
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	}
}

type clientMessageMap map[clientMessage][]chan clientMessage

func newClientMessageMap() clientMessageMap {
	m := make(clientMessageMap)

	for _, msg := range supportedMessages {
		m[msg] = make([]chan clientMessage, 0)
	}

	return m
}
