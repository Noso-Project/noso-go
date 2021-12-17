package common

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var (
	NOOP = func() {}
)

func newServer(t *testing.T, fn func()) (*httptest.Server, string) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn()
	}))

	// Get the IP:Port part of url
	addr, err := url.Parse(ts.URL)
	checkErr(t, err, nil)

	return ts, addr.Host
}

func newTestClient(addr string) *Client {
	client := NewClient(addr)
	client.connDialTimeout = 100 * time.Millisecond
	client.sendTimeout = 100 * time.Millisecond
	client.joinOkTimeout = 100 * time.Millisecond

	return client
}

func TestNewClient(t *testing.T) {
	t.Run("client.addr", func(t *testing.T) {
		got := NewClient("foo.com:1234").addr
		want := "foo.com:1234"

		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
	t.Run("client.connected", func(t *testing.T) {
		got := NewClient("foo.com:1234").connected
		want := false

		if got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	})
	t.Run("client.dialTimeout", func(t *testing.T) {
		got := NewClient("foo.com:1234").connDialTimeout
		want := connDialTimeout

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestConnect(t *testing.T) {
	t.Run("client connect", func(t *testing.T) {
		svr, addr := newServer(t, NOOP)
		defer svr.Close()

		client := newTestClient(addr)
		client.Connect()
		defer client.Disconnect()

		got := client.connected
		want := true

		if got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	})
	t.Run("client disconnect", func(t *testing.T) {
		svr, addr := newServer(t, NOOP)
		defer svr.Close()

		client := newTestClient(addr)
		client.Connect()
		client.Disconnect()

		got := client.connected
		want := false

		if got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	})
	t.Run("dial timeout", func(t *testing.T) {

		client := newTestClient("foo.com:1234")
		err := client.Connect()

		checkErr(t, err, DIALTIMEOUTERR)
	})
}

// JOINOK poolwalletaddr 0E0000000 PoolData 36808 0614B93EDBBD9151451072901C255792 11 0 104 0 -29 65009 3
const joinOkResp = "JOINOK poolwalletaddr minerprefix PoolData 36808 0614B93EDBBD9151451072901C255792 11 0 104 0 -29 65009 3\n"

func TestJoin(t *testing.T) {
	t.Run("join", func(t *testing.T) {
		cli, svr, done := getConns()
		defer close(done)

		go func() {
			rw := bufio.NewReadWriter(bufio.NewReader(svr), bufio.NewWriter(svr))
			for {
				select {
				case <-done:
					return
				default:
					req, _ := rw.ReadString('\n')
					fmt.Println("********************")
					fmt.Println("Received: ", req)
					fmt.Println("Sending : ", joinOkResp)
					fmt.Println("********************")
					rw.WriteString(joinOkResp)
					rw.Flush()
				}
			}

		}()

		client := newTestClient("fakeaddr.com:1234")
		client.conn = cli

		got := client.joined
		want := false

		if got != want {
			t.Fatalf("got %t, want %t", got, want)
		}

		subCh := client.Subscribe(JOINED)

		client.Connect()
		defer client.Disconnect()

		select {
		case msg := <-subCh:
			if msg != JOINED {
				t.Errorf("got clientMessage %q, want %d", msg, JOINED)
			}
			got = client.joined
			want = true

			if got != want {
				t.Errorf("got %t, want %t", got, want)
			}
		case <-time.After(5000 * time.Millisecond):
			t.Errorf("Timed out waiting for JOINED msg")
		}
	})
}

func getConns() (cli, svr net.Conn, done chan struct{}) {
	cli, svr = net.Pipe()
	done = make(chan struct{}, 0)

	go func() {
		defer cli.Close()
		defer svr.Close()

		<-done
	}()
	return
}

func checkErr(t *testing.T, got, want error) {
	t.Helper()

	if got != want {
		t.Fatalf("Got error %q, wanted error %q", got, want)
	}
}
