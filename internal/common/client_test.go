package common

import (
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
		// fmt.Println("In handler func")
		// fmt.Fprintln(w, "Hello, client")
	}))

	// Get the IP:Port part of url
	addr, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("got error and didn't expect one: %v", err)
	}

	return ts, addr.Host
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

		client := NewClient(addr)
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

		client := NewClient(addr)
		client.Connect()
		client.Disconnect()

		got := client.connected
		want := false

		if got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	})
	t.Run("dial timeout", func(t *testing.T) {

		client := NewClient("foo.com:1234")
		client.connDialTimeout = 1 * time.Nanosecond
		err := client.Connect()

		if err == nil {
			t.Errorf("expected a dial timeout but didn't get one")
		}
	})
}
