package common

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	ConnectTimeout          = 5 * time.Second
	DeadlineExceededTimeout = 15 * time.Second
	JoinTimeout             = 5 * time.Second
	PingInterval            = 5 * time.Second
)

var (
	ErrJoinTimeout = errors.New("Timed out while attempting to join pool")
	ErrPassFailed  = errors.New("Failed to join pool: wrong password")
)

func NewClient(done chan struct{}, poolAddr string, poolPort int) (client *Client) {
	// TODO: need to formalize done channels throughout
	// TODO: need to pass in poolPassword and walletAddress
	client = &Client{
		done:         done,
		poolAddr:     net.JoinHostPort(poolAddr, strconv.Itoa(poolPort)),
		auth:         "password leviable5",
		mu:           new(sync.Mutex),
		sendStream:   make(chan string, 0),
		joinTimeout:  JoinTimeout,
		pingInterval: PingInterval,
	}

	started := make(chan struct{}, 0)
	go client.start(started)

	<-started

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

func (c *Client) init() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = make(chan struct{}, 0)
	c.joined = make(chan struct{}, 0)
	c.broker = NewBroker(c.done)
}

func (c *Client) start(started chan struct{}) {
	var wg sync.WaitGroup

	for count := 0; ; count++ {
		c.init()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg.Add(3)
		go c.send(ctx, cancel, &wg)
		go c.recv(ctx, cancel, &wg)
		go c.ping(ctx, cancel, &wg)
		wg.Wait()

		if count > 0 {
			c.Connect()
		}

		// TODO: Use a sync.Once here
		select {
		case <-started:
		default:
			close(started)
		}

		select {
		case <-c.done:
			cancel()
			return
		case <-ctx.Done():
		}

		cancel()

		// TODO: Set configurable reconnect interval
		// fmt.Println("Will reconnect in 5 seconds")
		// time.Sleep(5 * time.Millisecond)
	}
}

func (c *Client) Connected() chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

func (c *Client) Joined() chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.joined
}

func (c *Client) Connect() (err error) {

	joinStream := c.broker.Subscribe(JoinTopic)
	defer c.broker.Unsubscribe(joinStream)

	if c.conn != nil {
		c.conn.Close()
	}

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
		return ErrJoinTimeout
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

// TODO: Really shouldn't use chan interface{} here
func (c *Client) Subscribe(topic Topic) <-chan interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.broker.Subscribe(topic)
}

// TODO: Really shouldn't use chan interface{} here
// TODO: Need to return an error here
func (c *Client) Unsubscribe(unsubStream <-chan interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.broker.Unsubscribe(unsubStream)
}

func (c *Client) send(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.sendStream:
			// fmt.Println("Send: ", msg)
			msg = c.auth + " " + msg
			fmt.Fprintln(c.conn, msg)
		}
	}
}

func (c *Client) recv(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	wg.Done()

	select {
	case <-c.connected:
	case <-ctx.Done():
		return
	}

	scanner := bufio.NewScanner(c.conn)

	for scanner.Scan() {
		c.conn.SetReadDeadline(time.Now().Add(DeadlineExceededTimeout))
		resp := scanner.Text()
		// fmt.Println("Recv: ", resp)
		msg, err := parse(resp)
		if err != nil {
			fmt.Println("Received an unknown response: ", resp)
		}
		// fmt.Println("Parsed msg: ", msg)
		c.broker.Publish(msg)
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			// TODO: Need to log that the deadline was exceeded and a
			//       reconnect attempt will happen
		} else {
			fmt.Fprintln(os.Stderr, "err reading scanner: ", err)
		}
	}

	cancel()
}

func (c *Client) ping(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	joinStream := c.Subscribe(JoinTopic)
	wg.Done()
	select {
	case <-ctx.Done():
		return
	case <-joinStream:
		c.Unsubscribe(joinStream)
	}

	ticker := time.NewTicker(c.pingInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// TODO: Need to use real hash rate value here instead of zero
			c.Send("PING 0")
		}
	}
}
