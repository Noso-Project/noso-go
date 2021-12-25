package common

// foo

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
)

const (
	DUMMYADDR = "fakeurl.com"
	DUMMYPORT = 12345
)

type Done chan struct{}

func NewTcpServer(t *testing.T) *TcpServer {
	svr := new(TcpServer)

	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("Caught an error and didn't expect one: %v", err)
	}

	svr.listener = l
	svr.addr = svr.listener.Addr().String()
	svr.done = make(Done, 0)

	host, port, err := net.SplitHostPort(svr.addr)

	if err != nil {
		t.Fatalf("Caught an error and didn't expect one: %v", err)
	}

	svr.host = host
	svr.port, _ = strconv.Atoi(port)

	var wg sync.WaitGroup
	wg.Add(1)
	go svr.Start(&wg)

	wg.Wait()

	return svr
}

type TcpServer struct {
	addr     string
	host     string
	port     int
	done     Done
	listener net.Listener
}

func (t *TcpServer) Start(wg *sync.WaitGroup) (err error) {
	wg.Done()
	// TODO: need to incorporate either a done channel or context
	for {

		conn, err := t.listener.Accept()
		if err != nil {
			err = errors.New("could not accept connection")
			break
		}
		if conn == nil {
			err = errors.New("could not create connection")
			break
		}

		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			// fmt.Println("Svr conn output: ", scanner.Text())
			fmt.Fprintln(conn, "JOINOK N6VxgLSpbni8kLbyUAjYXdHCPt2VEp 020000000 PoolData 37873 E1151A4F79E6394F6897A913ADCD476B 11 0 102 0 -30 42270 3")
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading connection: ", err)
		}

	}

	return
}

func (t *TcpServer) Close() (err error) {
	return t.listener.Close()
}

func TestClient(t *testing.T) {
	t.Run("new client", func(t *testing.T) {
		got := NewClient(make(chan struct{}, 0), DUMMYADDR, DUMMYPORT).poolAddr
		want := fmt.Sprintf("%s:%d", DUMMYADDR, DUMMYPORT)

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("connect", func(t *testing.T) {
		svr := NewTcpServer(t)
		defer svr.Close()

		client := NewClient(make(chan struct{}, 0), svr.host, svr.port)
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		got := client.connected
		want := true

		if got != want {
			t.Errorf("client.connected: got %t, wanted %t", got, want)
		}

		got = client.joined
		want = true

		if got != want {
			t.Errorf("client.joined: got %t, want %t", got, want)
		}
	})
}
