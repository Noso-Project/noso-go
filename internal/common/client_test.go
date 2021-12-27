package common

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	DUMMYADDR = "fakeurl.com"
	DUMMYPORT = 12345
)

func getClientSvr(t *testing.T) (*Client, *TcpServer, chan struct{}) {
	done := make(chan struct{}, 0)
	r := make(respMap)
	svr := NewTcpServer(done, t, r)
	client := NewClient(done, svr.Host, svr.Port)

	return client, svr, done
}

func TestClient(t *testing.T) {
	t.Run("new client", func(t *testing.T) {
		done := make(chan struct{}, 0)
		defer close(done)
		got := NewClient(done, DUMMYADDR, DUMMYPORT).poolAddr
		want := fmt.Sprintf("%s:%d", DUMMYADDR, DUMMYPORT)

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("connect refused", func(t *testing.T) {
		client, svr, done := getClientSvr(t)
		defer close(done)
		svr.Close()

		err := client.Connect()
		if err == nil {
			t.Fatal("Expected an error but didn't get one")
		}

		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("Expected 'connection refused' err, but got %s", err.Error())
		}
	})
	t.Run("connect timeout", func(t *testing.T) {
		oldTimeout := ConnectTimeout
		ConnectTimeout = 1 * time.Nanosecond
		defer func() { ConnectTimeout = oldTimeout }()

		client, svr, done := getClientSvr(t)
		defer close(done)
		svr.Close()

		err := client.Connect()
		if err == nil {
			t.Fatal("Expected an error but didn't get one")
		}

		if !strings.Contains(err.Error(), "i/o timeout") {
			t.Errorf("Expected 'connection refused' err, but got %s", err.Error())
		}
	})
	t.Run("connect successful", func(t *testing.T) {
		client, _, done := getClientSvr(t)
		defer close(done)

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		select {
		case <-client.connected:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client timeout out trying to connect")
		}
	})
	t.Run("join successful", func(t *testing.T) {
		client, _, done := getClientSvr(t)
		defer close(done)

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		select {
		case <-client.joined:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client timeout out trying to join")
		}
	})
	t.Run("join bad password", func(t *testing.T) {
		client, svr, done := getClientSvr(t)
		svr.rMap[JOIN] = []string{PASSFAILED_default}
		defer close(done)

		joinStream := client.broker.Subscribe(JoinTopic)
		client.Connect()

		select {
		case got := <-joinStream:
			switch got.(type) {
			case passFailed:
			default:
				t.Errorf("got %v, want passFailed", got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for server pong")
		}
	})
	t.Run("ping and pong", func(t *testing.T) {
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, _, done := getClientSvr(t)
		defer close(done)

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		pongStream := client.broker.Subscribe(PingPongTopic)
		defer close(pongStream)

		select {
		case <-pongStream:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for server pong")
		}
	})
	t.Run("poolsteps", func(t *testing.T) {
		// Use ping to trigger a poolsteps resp
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, svr, done := getClientSvr(t)
		defer close(done)

		svr.rMap[PING] = []string{POOLSTEPS_default}
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		poolStepsStream := client.broker.Subscribe(PoolStepsTopic)
		defer close(poolStepsStream)

		select {
		case <-poolStepsStream:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for server poolsteps")
		}
	})
	t.Run("pooldata from joinOk", func(t *testing.T) {
		client, _, done := getClientSvr(t)
		defer close(done)

		poolDataStream := client.broker.Subscribe(PoolDataTopic)
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		defer close(poolDataStream)

	loop:
		for {
			select {
			case msg := <-poolDataStream:
				switch msg.(type) {
				case joinOk:
					break loop
				default:
					continue
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timed out waiting for pooldata in joinOk msg")
			}
		}
	})
	t.Run("pooldata from pong", func(t *testing.T) {
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, _, done := getClientSvr(t)
		defer close(done)

		poolDataStream := client.broker.Subscribe(PoolDataTopic)
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		defer close(poolDataStream)
		after := time.After(100 * time.Millisecond)
	loop:
		for {
			select {
			case msg := <-poolDataStream:
				switch msg.(type) {
				case pong:
					break loop
				default:
					continue
				}
			case <-after:
				t.Fatal("Timed out waiting for pooldata in pong msg")
			}
		}
	})
	t.Run("pooldata from poolSteps", func(t *testing.T) {
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, svr, done := getClientSvr(t)
		defer close(done)

		svr.rMap[PING] = []string{POOLSTEPS_default}

		poolDataStream := client.broker.Subscribe(PoolDataTopic)
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		defer close(poolDataStream)
		after := time.After(100 * time.Millisecond)
	loop:
		for {
			select {
			case msg := <-poolDataStream:
				switch msg.(type) {
				case poolSteps:
					break loop
				default:
					continue
				}
			case <-after:
				t.Fatal("Timed out waiting for pooldata in poolSteps msg")
			}
		}
	})
	t.Run("stepok", func(t *testing.T) {
		// Use ping to trigger a stepok resp
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, svr, done := getClientSvr(t)
		defer close(done)

		svr.rMap[PING] = []string{STEPOK_default}
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		stepOkStream := client.broker.Subscribe(StepOkTopic)
		defer close(stepOkStream)

		select {
		case resp := <-stepOkStream:
			switch resp.(type) {
			case stepOk:
				got := resp.(stepOk).PopValue
				want := 256

				if got != want {
					t.Errorf("got %d, want %d", got, want)
				}
			default:
				t.Errorf("Expected stepOk msg, but got %v", resp)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for server StepOk")
		}
	})
}
