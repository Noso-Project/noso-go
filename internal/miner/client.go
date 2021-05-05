package miner

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type managerComms struct {
	connected    chan struct{}
	disconnected chan struct{}
	stop         chan struct{}
	joined       chan struct{}
	sendStopped  chan struct{}
	recvStopped  chan struct{}
	pingStopped  chan struct{}
	quit         chan struct{}
}

func NewManagerComms() *managerComms {
	return &managerComms{
		connected:    make(chan struct{}, 0),
		disconnected: make(chan struct{}, 0),
		stop:         make(chan struct{}, 0),
		joined:       make(chan struct{}, 0),
		sendStopped:  make(chan struct{}, 0),
		recvStopped:  make(chan struct{}, 0),
		pingStopped:  make(chan struct{}, 0),
		quit:         make(chan struct{}, 0),
	}
}

func NewTcpClient(opts *Opts, comms *Comms) *TcpClient {
	client := &TcpClient{
		minerVer:  MinerName,
		comms:     comms,
		addr:      fmt.Sprintf("%s:%d", opts.IpAddr, opts.IpPort),
		auth:      fmt.Sprintf("%s %s", opts.PoolPw, opts.Wallet),
		SendChan:  make(chan string, 0),
		RecvChan:  make(chan string, 0),
		connected: make(chan interface{}, 0),
	}

	go client.manager()

	return client
}

type TcpClient struct {
	minerVer  string
	comms     *Comms
	addr      string // "poolIP:poolPort"
	auth      string // "poolPw wallet"
	SendChan  chan string
	RecvChan  chan string
	conn      net.Conn
	connected chan interface{}
	manComms  *managerComms
}

func (t *TcpClient) Close() {
	close(t.manComms.quit)
}

// Manages the TCP connection and send/recv/ping goroutines
func (t *TcpClient) manager() {
	for {
		t.conn = nil
		t.manComms = NewManagerComms()

		go t.send()
		go t.recv()
		go t.ping()

		for running := true; running; {
			select {
			case <-t.manComms.disconnected:
				close(t.manComms.stop)
				running = false
			case <-t.comms.Joined:
				close(t.manComms.joined)
			}
		}

		// Wait for the goroutines to exit
		// Should probably use a waitgroup here?
		<-t.manComms.sendStopped
		<-t.manComms.recvStopped
		<-t.manComms.pingStopped

		// Wait 5 seconds between connection attempts
		time.Sleep(5 * time.Second)
	}
}

func (t *TcpClient) send() {
	var (
		err error
	)
	t.conn, err = net.DialTimeout("tcp", t.addr, 5*time.Second)
	if err != nil {
		t.conn = nil
		fmt.Printf("Error connecting to pool: %v\n", err)
		t.Close()
		close(t.manComms.sendStopped)
		return

	} else {
		t.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
		close(t.manComms.connected)
	}

	go t.join()

	for {
		select {
		case msg := <-t.SendChan:
			fmt.Printf("-> %s\n", msg)
			msg = fmt.Sprintf("%s %s\n", t.auth, msg)
			fmt.Fprintf(t.conn, msg)
		case <-t.manComms.stop:
			close(t.manComms.sendStopped)
			return
		}
	}
}

func (t *TcpClient) recv() {
	// Block until connection established
	select {
	case <-t.manComms.connected:
	case <-t.manComms.stop:
		close(t.manComms.recvStopped)
		return
	}
	<-t.manComms.connected
	scanner := bufio.NewScanner(t.conn)
	for connected := true; connected; {
		if ok := scanner.Scan(); !ok {
			fmt.Println("Error in connection: ", scanner.Err())
			connected = false
		}
		resp := scanner.Text()
		if resp == "" {
			continue
		}
		fmt.Print("<- " + resp + "\n")
		t.RecvChan <- resp
		// Since we got something, reset the deadline
		t.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	}
	close(t.manComms.disconnected)
	close(t.manComms.recvStopped)
	return
}

func (t *TcpClient) ping() {
	var hashRate int

	go func() {
		for {
			select {
			case hashRate = <-t.comms.HashRate:
			case <-t.manComms.stop:
				return
			}
		}
	}()

	// Block until connected to pool
	select {
	case <-t.manComms.connected:
	case <-t.manComms.stop:
		close(t.manComms.pingStopped)
		return
	}

	// Block until pool has been joined
	select {
	case <-t.manComms.joined:
	case <-t.manComms.stop:
		close(t.manComms.pingStopped)
		return
	}

	for connected := true; connected; {
		select {
		case <-t.manComms.stop:
			connected = false
		case <-time.After(5 * time.Second):
			t.SendChan <- fmt.Sprintf("PING %d", hashRate/1000)
		}
	}
	close(t.manComms.pingStopped)
}

func (t *TcpClient) join() {
	t.SendChan <- fmt.Sprintf("JOIN %s", t.minerVer)
}
