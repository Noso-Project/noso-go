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
	// logger *zap.SugaredLogger

	// Timeouts and Interval timers
	ConnectTimeout          = 5 * time.Second
	DeadlineExceededTimeout = 15 * time.Second
	JoinTimeout             = 5 * time.Second
	PingInterval            = 5 * time.Second
	ReconnectWait           = 5 * time.Second
)

var (
	ErrJoinTimeout      = errors.New("Timed out while attempting to join pool")
	ErrPassFailed       = errors.New("Failed to join pool: wrong password")
	ErrAlreadyConnected = errors.New("Failed to join pool: already connected")
)

func NewClient(ctx context.Context, poolAddr string, poolPort int) (client *Client) {
	// TODO: need to formalize done channels throughout
	// TODO: need to pass in poolPassword and walletAddress
	InitLogger(os.Stdout)
	client = &Client{
		parentCtx:       ctx,
		poolAddr:        net.JoinHostPort(poolAddr, strconv.Itoa(poolPort)),
		auth:            "password leviable6",
		mu:              new(sync.Mutex),
		sendStream:      make(chan string, 0),
		doConnect:       make(chan struct{}),
		doConnectErr:    make(chan error),
		connTimeout:     ConnectTimeout,
		deadlineTimeout: DeadlineExceededTimeout,
		joinTimeout:     JoinTimeout,
		pingInterval:    PingInterval,
	}

	started := make(chan struct{}, 0)
	logger.Debug("Client starting")
	go client.start(started, true)

	// TODO: Need to protect this with select and return err
	<-started
	logger.Debug("Client started")

	return client
}

func NewClientWithConn(ctx context.Context, conn net.Conn) (client *Client) {
	InitLogger(os.Stdout)

	client = &Client{
		parentCtx:       ctx,
		poolAddr:        net.JoinHostPort("", strconv.Itoa(0)),
		auth:            "",
		conn:            conn,
		mu:              new(sync.Mutex),
		sendStream:      make(chan string, 0),
		doConnect:       make(chan struct{}),
		doConnectErr:    make(chan error),
		connTimeout:     ConnectTimeout,
		deadlineTimeout: DeadlineExceededTimeout,
		joinTimeout:     JoinTimeout,
		pingInterval:    PingInterval,
	}

	started := make(chan struct{}, 0)
	logger.Debug("Client starting")
	go client.start(started, false)

	// TODO: Need to protect this with select and return err
	<-started

	close(client.connected)
	close(client.joined)

	logger.Debug("Client started")

	return client
}

type Client struct {
	// TODO: Evaluate using a context instead of done channel
	// TODO: Set auth
	parentCtx    context.Context
	poolAddr     string
	auth         string // "poolPw walletAddr"
	conn         net.Conn
	connected    chan struct{}
	doConnect    chan struct{}
	doConnectErr chan error
	joined       chan struct{}
	sendStream   chan string
	broker       *Broker
	mu           *sync.Mutex

	// Timeouts/Intervals
	connTimeout     time.Duration
	deadlineTimeout time.Duration
	joinTimeout     time.Duration
	pingInterval    time.Duration
}

func (c *Client) init() (context.Context, context.CancelFunc) {
	logger.Debug("Initializing client")
	c.mu.Lock()
	logger.Debug("init() has the lock")
	defer logger.Debug("init() released the lock")
	defer c.mu.Unlock()
	ctx, cancel := context.WithCancel(c.parentCtx)
	c.connected = make(chan struct{}, 0)
	c.joined = make(chan struct{}, 0)
	c.broker = NewBroker(ctx, cancel)

	logger.Debug("Client initialized")

	return ctx, cancel
}

func (c *Client) start(started chan struct{}, withConn bool) {
	var wg sync.WaitGroup
	var once sync.Once

	for count := 0; ; count++ {
		logger.Debugf("Count is: %d", count)
		ctx, cancel := c.init()
		// TODO: need to wrap this section in func so cancel can be called
		//       and garbage collected
		defer func(cancel context.CancelFunc) { logger.Debugf("Calling deferred cancel()"); cancel() }(cancel)

		logger.Debugf("Starting send, recv, and ping goroutines")
		wg.Add(3)
		go c.send(ctx, cancel, &wg)
		go c.recv(ctx, cancel, &wg)
		go c.ping(ctx, cancel, &wg)
		wg.Wait()
		logger.Debugf("send, recv, and ping started")

		once.Do(func() { close(started) })

		if withConn == true {
			// Wait for trigger to connect from Connect() method
			<-c.doConnect
			err := c.connect()
			if err != nil {
				switch err.(error) {
				case ErrAlreadyConnected:
					cancel()
					continue
				default:
					select {
					case c.doConnectErr <- err:
					default:
						logger.Panic(err)
					}
				}
			} else {
				select {
				case c.doConnectErr <- nil:
				default:
				}
			}
		}

		select {
		case <-c.parentCtx.Done():
			logger.Debug("<-parentCtx.Done() closed")
			cancel()
			return
		case <-ctx.Done():
			logger.Debug("<-ctx.Done() closed")
		}

		// TODO: Set configurable reconnect interval
		logger.Infof("Will reconnect in %s seconds", ReconnectWait)
		time.Sleep(ReconnectWait)
		logger.Info("About to reconnect")
	}
}

func (c *Client) Connected() chan struct{} {
	c.mu.Lock()
	logger.Debug("Connected() has the lock")
	defer logger.Debug("Connected() released the lock")
	defer c.mu.Unlock()
	return c.connected
}

func (c *Client) Joined() chan struct{} {
	c.mu.Lock()
	logger.Debug("Joined() has the lock")
	defer logger.Debug("Joined() released the lock")
	defer c.mu.Unlock()
	return c.joined
}
func (c *Client) Connect() error {
	close(c.doConnect)
	return <-c.doConnectErr
}

func (c *Client) connect() (err error) {

	if c.conn != nil {
		logger.Debug("Found existing connection; closing")
		c.conn.Close()
	}

	c.conn, err = net.DialTimeout("tcp", c.poolAddr, c.connTimeout)
	if err != nil {
		return err
	}

	close(c.connected)

	logger.Debug("Subscribing to JoinTopic")
	joinStream, err := c.Subscribe(JoinTopic)
	if err != nil {
		return err
	}
	logger.Debug("Subscribed to JoinTopic for stream: ", joinStream)
	defer logger.Debug("deferred Unsubscribe for ", joinStream)
	defer c.Unsubscribe(joinStream)

	// TODO: Might need to explicitely separate connect and join,
	//       as its possible a secondary client might want to connect,
	//       to a pool but not join it until the primary fails
	c.join()

	select {
	case resp := <-joinStream:
		logger.Debugf("JOIN resp: %s", resp)
		switch resp.(type) {
		case JoinOk:
		case AlreadyConnected:
			return ErrAlreadyConnected
		default:
		}
	case <-time.After(c.joinTimeout):
		return ErrJoinTimeout
	}

	close(c.joined)
	logger.Debug("c.connect() complete")
	return nil
}

func (c *Client) join() {
	// TODO: Need to use real values for version and instanceId
	c.Send("JOIN ng9.9.9")
}

// TODO: Pass in Tx objects here instead of strings
func (c *Client) Send(msg string) {
	go func(msg string) {
		// TODO: Should I make this timeout? Or use a context deadline?
		select {
		case <-c.parentCtx.Done():
			logger.Debug("<-c.parentCtx.Done() closed")
			return
		case c.sendStream <- msg:
			logger.Debugf("Sent %s to %v", msg, c.sendStream)
		}
	}(msg)
}

// TODO: Really shouldn't use chan interface{} here
func (c *Client) Publish(pub interface{}) {
	// TODO: There is an expectation here that Subscribe blocks
	//       until we are actually subscribed
	c.mu.Lock()
	logger.Debug("Publish() has the lock")
	defer logger.Debug("Publish() released the lock")
	defer c.mu.Unlock()
	c.broker.Publish(pub)
}

// TODO: Really shouldn't use chan interface{} here
func (c *Client) Subscribe(topic Topic) (<-chan interface{}, error) {
	// TODO: There is an expectation here that Subscribe blocks
	//       until we are actually subscribed
	c.mu.Lock()
	logger.Debug("Subscribe() has the lock")
	defer logger.Debug("Subscribe() released the lock")
	defer c.mu.Unlock()
	return c.broker.Subscribe(topic)
}

// TODO: Really shouldn't use chan interface{} here
// TODO: Need to return an error here
func (c *Client) Unsubscribe(unsubStream <-chan interface{}) {
	c.mu.Lock()
	logger.Debug("Unsubscribe() has the lock")
	defer logger.Debug("Unsubscribe() released the lock")
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
			logger.Debug("Send: ", msg)
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

	deadline := c.deadlineTimeout
	for scanner.Scan() {
		c.conn.SetReadDeadline(time.Now().Add(deadline))
		resp := scanner.Text()
		logger.Debug("Recv: ", resp)
		msg, err := parse(resp)
		if err != nil {
			logger.Error("Received an unknown response: ", resp)
		}
		// logger.Debug("Parsed msg: ", msg)
		c.broker.Publish(msg)
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			// TODO: Need to log that the deadline was exceeded and a
			//       reconnect attempt will happen
			logger.Error("deadline exceeded: ", err)
		} else {
			logger.Error("err reading scanner: ", err)
		}
	}

	cancel()
}

func (c *Client) ping(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	wg.Done()
	select {
	case <-ctx.Done():
		return
	case <-c.Joined():
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
