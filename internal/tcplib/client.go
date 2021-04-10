package tcplib

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/leviable/noso-go/internal/miner"
)

func NewTcpClient(opts *miner.Opts) *TcpClient {
	client := &TcpClient{
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
	addr      string // "poolIP:poolPort"
	auth      string // "poolPw wallet"
	SendChan  chan string
	RecvChan  chan string
	conn      net.Conn
	connected chan interface{}
}

func (t *TcpClient) send() {
	if t.conn == nil {
		t.conn, _ = net.DialTimeout("tcp", t.addr, 5*time.Second)
		// TODO: figure out what to do on error
		// TODO: figure out reconnecting
		// if err != nil {
		// 	return err
		// }
		close(t.connected)
	}

	for {
		select {
		case msg := <-t.SendChan:
			msg = fmt.Sprintf("%s %s\n", t.auth, msg)
			fmt.Printf("-> %s", msg)
			fmt.Fprintf(t.conn, msg)
		}
	}
}

func (t *TcpClient) recv() {
	// Block until connection established
	<-t.connected
	scanner := bufio.NewScanner(t.conn)
	for scanner.Scan() {
		resp := scanner.Text()
		fmt.Print("<- " + resp + "\n")
		t.RecvChan <- resp
	}
}

func (t *TcpClient) ping() {
	// Block until connection established
	<-t.connected
	for {
		select {
		case <-time.After(5 * time.Second):
			t.SendChan <- "PING 1"
		}
	}
	scanner := bufio.NewScanner(t.conn)
	for scanner.Scan() {
		resp := scanner.Text()
		fmt.Print("<- " + resp + "\n")
		t.RecvChan <- resp
	}
}
