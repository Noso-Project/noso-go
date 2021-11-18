package miner

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	dialTimeout       = 5 * time.Second
	connectionTimeout = 20 * time.Second
	reconnectSleep    = 5 * time.Second
)

func NewTcpClient(opts *Opts, comms *Comms, showLogs, join bool) *TcpClient {
	client := &TcpClient{
		minerVer:    MinerName,
		comms:       comms,
		opts:        opts,
		addr:        fmt.Sprintf("%s:%d", opts.IpAddr, opts.IpPort),
		SendChan:    make(chan string, 100),
		RecvChan:    make(chan string, 100),
		connected:   make(chan interface{}, 0),
		mutex:       &sync.Mutex{},
		showLogs:    showLogs,
		join:        join,
		exitOnRetry: opts.ExitOnRetry,
	}

	go client.manager()

	return client
}

type TcpClient struct {
	minerVer    string
	comms       *Comms
	opts        *Opts
	addr        string // "poolIP:poolPort"
	auth        string // "poolPw wallet"
	SendChan    chan string
	RecvChan    chan string
	conn        net.Conn
	connected   chan interface{}
	mutex       *sync.Mutex
	showLogs    bool
	join        bool
	exitOnRetry bool
}

type managerComms struct {
	connected    chan struct{}
	disconnected chan struct{}
	joined       chan struct{}
}

func NewManagerComms() *managerComms {
	return &managerComms{
		connected:    make(chan struct{}, 0),
		disconnected: make(chan struct{}, 0),
		joined:       make(chan struct{}, 0),
	}
}

func (t *TcpClient) SetAuth() {
	t.opts.CurrentWallet = t.opts.Wallets[0]
	t.opts.Wallets = append(t.opts.Wallets[1:], t.opts.CurrentWallet)
	t.auth = fmt.Sprintf("%s %s", t.opts.PoolPw, t.opts.CurrentWallet)

	log.Printf("Using wallet address: %s\n", t.opts.CurrentWallet)
}

// Manages the TCP connection and send/recv/ping goroutines
func (t *TcpClient) manager() {
	for {
		manComms := NewManagerComms()
		t.comms.Disconnected = manComms.disconnected

		conn, err := net.DialTimeout("tcp", t.addr, dialTimeout)
		if err != nil {
			log.Printf("Error connecting to pool: %v\n", err)
		} else {
			conn.SetReadDeadline(time.Now().Add(connectionTimeout))

			go t.send(conn, manComms)
			go t.recv(conn, manComms)
			go t.ping(manComms)
			go t.watchDog(manComms)

		manager:
			for {
				select {
				case <-manComms.disconnected:
					break manager
				case <-t.comms.Joined:
					t.close(manComms.joined)
				}
			}

			conn.Close()
		}

		if t.join && !t.exitOnRetry {
			// Wait 5 seconds between connection attempts
			log.Printf("Disconnected from pool, will retry connection in %d seconds\n", reconnectSleep/time.Second)
			time.Sleep(reconnectSleep)
		}
	}
}

func (t *TcpClient) send(conn net.Conn, manComms *managerComms) {
	if t.join {
		go func() { t.SendChan <- fmt.Sprintf("JOIN %s", t.minerVer) }()
	}

send:
	for {
		select {
		case msg := <-t.SendChan:
			if msg[:4] == "JOIN" {
				t.SetAuth()
			}

			if t.showLogs {
				log.Printf("-> %s\n", msg)
			}
			msg = fmt.Sprintf("%s %s\n", t.auth, msg)
			fmt.Fprintf(conn, msg)
		case <-manComms.disconnected:
			break send
		}
	}
}

func (t *TcpClient) recv(conn net.Conn, manComms *managerComms) {
	scanner := bufio.NewScanner(conn)
recv:
	for {
		select {
		case <-manComms.disconnected:
			break recv
		default:
			if ok := scanner.Scan(); !ok {
				if t.showLogs {
					log.Println("Error in connection: ", scanner.Err())
				}
				t.close(manComms.disconnected)
				break
			}
			resp := scanner.Text()
			if resp == "" {
				continue
			}
			if t.showLogs {
				log.Print("<- " + resp + "\n")
			}
			t.RecvChan <- resp
			// Since we got something, reset the deadline
			conn.SetReadDeadline(time.Now().Add(connectionTimeout))
		}
	}
}

func (t *TcpClient) ping(manComms *managerComms) {
	var (
		hashRate int
		m        sync.RWMutex
	)

	go func() {
		for {
			select {
			case <-manComms.disconnected:
				return
			case hr := <-t.comms.HashRate:
				m.Lock()
				hashRate = hr
				m.Unlock()
			}
		}
	}()

	// Block until pool has been joined
	select {
	case <-manComms.joined:
	case <-manComms.disconnected:
		return
	}

ping:
	for {
		select {
		case <-manComms.disconnected:
			break ping
		case <-time.After(5 * time.Second):

			m.RLock()
			hr := hashRate
			m.RUnlock()
			t.SendChan <- fmt.Sprintf("PING %d", hr/1000)
		}
	}
}

func (t *TcpClient) watchDog(manComms *managerComms) {
	// If we don't get a PONG back after X seconds, reconnect
watchdog:
	for {
		select {
		case <-t.comms.Pong:
			continue
		case <-manComms.disconnected:
			break watchdog
		case <-time.After(connectionTimeout):
			log.Printf("###################\nWatchdog Triggered\n###################\n")
			t.close(manComms.disconnected)
			break watchdog
		}
	}
}

func (t *TcpClient) close(c chan struct{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	select {
	case <-c:
	default:
		close(c)
	}
}
