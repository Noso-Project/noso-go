package miner

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func NewTcpClient(opts *Opts, comms *Comms) *TcpClient {
	client := &TcpClient{
		minerVer:  MinerName,
		comms:     comms,
		addr:      fmt.Sprintf("%s:%d", opts.IpAddr, opts.IpPort),
		auth:      fmt.Sprintf("%s %s", opts.PoolPw, opts.Wallet),
		SendChan:  make(chan string, 10),
		RecvChan:  make(chan string, 10),
		connected: make(chan interface{}, 10),
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
}

type managerComms struct {
	connected    chan struct{}
	disconnected chan struct{}
	stop         chan struct{}
	joined       chan struct{}
	sendStopped  chan struct{}
	recvStopped  chan struct{}
	pingStopped  chan struct{}
}

func NewManagerComms() *managerComms {
	return &managerComms{
		connected:    make(chan struct{}, 10),
		disconnected: make(chan struct{}, 10),
		stop:         make(chan struct{}, 10),
		joined:       make(chan struct{}, 10),
		sendStopped:  make(chan struct{}, 10),
		recvStopped:  make(chan struct{}, 10),
		pingStopped:  make(chan struct{}, 10),
	}
}

// Manages the TCP connection and send/recv/ping goroutines
func (t *TcpClient) manager() {
	for {
		t.conn = nil
		manComms := NewManagerComms()

		go t.send(manComms)
		go t.recv(manComms)
		go t.ping(manComms)

		for running := true; running; {
			select {
			case <-manComms.disconnected:
				close(manComms.stop)
				running = false
			case <-t.comms.Joined:
				close(manComms.joined)
			}
		}

		// Wait for the goroutines to exit
		// Should probably use a waitgroup here?
		<-manComms.sendStopped
		<-manComms.recvStopped
		<-manComms.pingStopped

		// Wait 5 seconds between connection attempts
		time.Sleep(5 * time.Second)
	}
}

func (t *TcpClient) send(manComms *managerComms) {
	var (
		err error
	)
	t.conn, err = net.DialTimeout("tcp", t.addr, 5*time.Second)
	if err != nil {
		t.conn = nil
		fmt.Printf("Error connecting to pool: %v\n", err)
		close(manComms.disconnected)
		close(manComms.sendStopped)
		return

	} else {
		t.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
		close(manComms.connected)
	}

	go t.join()

	for {
		select {
		case msg := <-t.SendChan:
			fmt.Printf("-> %s\n", msg)
			msg = fmt.Sprintf("%s %s\n", t.auth, msg)
			fmt.Fprintf(t.conn, msg)
		case <-manComms.stop:
			close(manComms.sendStopped)
			return
		}
	}
}

func (t *TcpClient) recv(manComms *managerComms) {
	// Block until connection established
	select {
	case <-manComms.connected:
	case <-manComms.stop:
		close(manComms.recvStopped)
		return
	}
	<-manComms.connected
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
	close(manComms.disconnected)
	close(manComms.recvStopped)
	return
}

func (t *TcpClient) ping(manComms *managerComms) {
	var hashRate int

	go func() {
		for {
			select {
			case hashRate = <-t.comms.HashRate:
			case <-manComms.stop:
				return
			}
		}
	}()

	// Block until connected to pool
	select {
	case <-manComms.connected:
	case <-manComms.stop:
		close(manComms.pingStopped)
		return
	}

	// Block until pool has been joined
	select {
	case <-manComms.joined:
	case <-manComms.stop:
		close(manComms.pingStopped)
		return
	}

	for connected := true; connected; {
		select {
		case <-manComms.stop:
			connected = false
		case <-time.After(5 * time.Second):
			t.SendChan <- fmt.Sprintf("PING %d", hashRate/1000)
		}
	}
	close(manComms.pingStopped)
}

func (t *TcpClient) join() {
	t.SendChan <- fmt.Sprintf("JOIN %s", t.minerVer)
}
