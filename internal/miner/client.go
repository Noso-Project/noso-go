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
		connected:    make(chan struct{}, 0),
		disconnected: make(chan struct{}, 0),
		stop:         make(chan struct{}, 0),
		joined:       make(chan struct{}, 0),
		sendStopped:  make(chan struct{}, 0),
		recvStopped:  make(chan struct{}, 0),
		pingStopped:  make(chan struct{}, 0),
	}
}

// Manages the TCP connection and send/recv/ping goroutines
func (t *TcpClient) manager() {
	for {
		t.conn = nil
		t.manComms = NewManagerComms()

		go t.send()
		go t.recv()
		go t.ping()
		go t.watchDog()

		for running := true; running; {
			select {
			case <-t.manComms.disconnected:
				t.close(t.manComms.stop)
				running = false
			case <-t.comms.Joined:
				t.close(t.manComms.joined)
			}
		}

		// Wait for the goroutines to exit
		// Should probably use a waitgroup here?
		<-t.manComms.sendStopped
		<-t.manComms.recvStopped
		<-t.manComms.pingStopped

		// Wait 5 seconds between connection attempts
		time.Sleep(500 * time.Millisecond)
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
		t.close(t.manComms.disconnected)
		t.close(t.manComms.sendStopped)
		return

	} else {
		t.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
		t.close(t.manComms.connected)
	}

	go t.join()

	for {
		select {
		case msg := <-t.SendChan:
			fmt.Printf("-> %s\n", msg)
			msg = fmt.Sprintf("%s %s\n", t.auth, msg)
			fmt.Fprintf(t.conn, msg)
		case <-t.manComms.stop:
			t.close(t.manComms.sendStopped)
			return
		}
	}
}

func (t *TcpClient) recv() {
	// Block until connection established
	select {
	case <-t.manComms.connected:
	case <-t.manComms.stop:
		t.close(t.manComms.recvStopped)
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
	t.close(t.manComms.disconnected)
	t.close(t.manComms.recvStopped)
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
		t.close(t.manComms.pingStopped)
		return
	}

	// Block until pool has been joined
	select {
	case <-t.manComms.joined:
	case <-t.manComms.stop:
		t.close(t.manComms.pingStopped)
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
	t.close(t.manComms.pingStopped)
}

func (t *TcpClient) join() {
	t.SendChan <- fmt.Sprintf("JOIN %s", t.minerVer)
}

func (t *TcpClient) close(c chan struct{}) {
	select {
	case <-c:
		//fmt.Printf("********* Chan %+v already closed\n", c)
	default:
		//fmt.Printf("********* Closing Chan %+v \n", c)
		close(c)
	}
}

func (t *TcpClient) watchDog() {
	// If we don't get a PONG back after X seconds, reconnect
	for {
		select {
		case <-t.comms.Pong:
			continue
		case <-t.manComms.disconnected:
			break
		case <-time.After(20 * time.Second):
			fmt.Printf("###################\nWatchdog Triggered\n###################\n")
			t.close(t.manComms.disconnected)
			break
		}
	}
}
