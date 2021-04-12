package miner

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func NewTcpClient(opts *Opts, comms *Comms) *TcpClient {
	client := &TcpClient{
		comms:     comms,
		addr:      fmt.Sprintf("%s:%d", opts.IpAddr, opts.IpPort),
		auth:      fmt.Sprintf("%s %s", opts.PoolPw, opts.Wallet),
		SendChan:  make(chan string, 0),
		RecvChan:  make(chan string, 0),
		connected: make(chan interface{}, 0),
	}

	go client.send()
	go client.recv()
	go client.ping()

	return client
}

type TcpClient struct {
	comms     *Comms
	addr      string // "poolIP:poolPort"
	auth      string // "poolPw wallet"
	SendChan  chan string
	RecvChan  chan string
	conn      net.Conn
	connected chan interface{}
}

func (t *TcpClient) send() {
	var (
		err         error
		lastPong    time.Time
		pongTimeout time.Duration
	)
	lastPong = time.Now()
	for {
		if t.conn == nil {
			t.comms.Joined = make(chan interface{}, 0)
			t.conn, err = net.DialTimeout("tcp", t.addr, 5*time.Second)
			if err != nil {
				t.conn = nil
				fmt.Printf("Error connecting to pool: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			} else {
				go t.join()
			}
			//else {
			//	t.SendChan <- fmt.Sprintf("JOIN foo")
			//}
			close(t.connected)
			lastPong = time.Now()
		}

		// Timeout if we haven't received a PONG in 20 seconds
		pongTimeout = lastPong.Add(20 * time.Second).Sub(time.Now())
		select {
		case msg := <-t.SendChan:
			fmt.Printf("-> %s\n", msg)
			fmt.Printf("PONG timeout %s\n", pongTimeout.String())
			msg = fmt.Sprintf("%s %s\n", t.auth, msg)
			fmt.Fprintf(t.conn, msg)
		case <-t.comms.Pong:
			lastPong = time.Now()
		case <-time.After(pongTimeout):
			// connection died, try to reconnect
			fmt.Println("Connection died, attempting to reconnect")
			t.conn = nil
			t.connected = make(chan interface{}, 0)
		}
	}
}

func (t *TcpClient) recv() {
	for {
		// Block until connection established
		if t.conn != nil {
			<-t.connected
			scanner := bufio.NewScanner(t.conn)
			if scanner.Scan() {
				resp := scanner.Text()
				fmt.Print("<- " + resp + "\n")
				t.RecvChan <- resp
			}
		}
	}
}

func (t *TcpClient) ping() {
	var hashRate int

	go func() {
		for {
			select {
			case hashRate = <-t.comms.HashRate:
			}
		}
	}()

	for {
		// Block until connected to pool and joined
		<-t.connected
		<-t.comms.Joined
		select {
		case <-time.After(5 * time.Second):
			t.SendChan <- fmt.Sprintf("PING %d", hashRate/1000)
		}
	}
}

func (t *TcpClient) join() {
	t.SendChan <- fmt.Sprintf("JOIN %s", "foo")
}
