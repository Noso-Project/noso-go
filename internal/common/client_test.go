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

func getClientSvr(t testing.TB) (*Client, *TcpServer, chan struct{}) {
	done := make(chan struct{}, 0)
	r := make(respMap)
	svr := NewTcpServer(done, t, r)
	client := NewClient(done, svr.Host, svr.Port)

	return client, svr, done
}

func TestClientConnect(t *testing.T) {
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
		case <-client.Connected():
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
		case <-client.Joined():
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client timeout out trying to join")
		}
	})
	t.Run("join bad password", func(t *testing.T) {
		client, svr, done := getClientSvr(t)
		svr.rMap[JOIN] = []string{PASSFAILED_default}
		defer close(done)

		joinStream := client.Subscribe(JoinTopic)
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
	t.Run("reconnect on closed connection", func(t *testing.T) {
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, svr, done := getClientSvr(t)
		defer close(done)

		svr.printConnErr = false

		pingStream := client.Subscribe(PingPongTopic)
		client.Connect()

		// Get one pong, close the svr conn, then
		// wait for one more pong

		select {
		case <-pingStream:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for server pong")
		}

		client.Unsubscribe(pingStream)

		// Subtle race condition here:
		//  - After closing the connection, its possible that
		//  - the Connected() and Joined() checks will work
		//  - because they are the old, already closed channels,
		//  - and then the Subscribe call goes to the old broker
		//
		// Need to wait for the client to re-init before checking
		// its connected/joined status

		oldConnectedChan := client.Connected()

		svr.conn.Close()

		for oldConnectedChan == client.Connected() {
			fmt.Println("Still the old chan")
			time.Sleep(100 * time.Microsecond)
		}

		// Wait for connect
		select {
		case <-client.Connected():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting to connect to pool")
		}

		// Wait for join
		select {
		case <-client.Joined():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting to rejoin pool")
		}

		pingStream = client.Subscribe(PingPongTopic)

		select {
		case <-pingStream:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for server pong")
		}
	})
	// t.Run("reconnect on read deadline exceeded", func(t *testing.T) {
	// 	oldPing := PingInterval
	// 	PingInterval = 10 * time.Millisecond
	// 	defer func() { PingInterval = oldPing }()

	// 	client, svr, done := getClientSvr(t)
	// 	defer close(done)

	// 	svr.printConnErr = false

	// 	pingStream := client.Subscribe(PingPongTopic)
	// 	client.Connect()

	// 	// Get one pong, close the svr conn, then
	// 	// wait for one more pong

	// 	select {
	// 	case <-pingStream:
	// 	case <-time.After(100 * time.Millisecond):
	// 		t.Fatal("Timed out waiting for server pong")
	// 	}

	// 	client.Unsubscribe(pingStream)

	// 	svr.conn.Close()

	// 	// Wait for connect
	// 	select {
	// 	case <-client.Connected():
	// 	case <-time.After(100 * time.Millisecond):
	// 		t.Fatal("Timed out waiting to connect to pool")
	// 	}

	// 	// Wait for join
	// 	select {
	// 	case <-client.Joined():
	// 	case <-time.After(100 * time.Millisecond):
	// 		t.Fatal("Timed out waiting to rejoin pool")
	// 	}

	// 	pingStream = client.Subscribe(PingPongTopic)

	// 	select {
	// 	case <-pingStream:
	// 	case <-time.After(500 * time.Millisecond):
	// 		t.Fatal("Timed out waiting for server pong")
	// 	}
	// })
}

func TestClientMessaging(t *testing.T) {
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

		pongStream := client.Subscribe(PingPongTopic)
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

		poolStepsStream := client.Subscribe(PoolStepsTopic)
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

		poolDataStream := client.Subscribe(PoolDataTopic)
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

		poolDataStream := client.Subscribe(PoolDataTopic)
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

		poolDataStream := client.Subscribe(PoolDataTopic)
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

		stepOkStream := client.Subscribe(StepOkTopic)
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

func BenchmarkSend(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	client, svr, done := getClientSvr(b)
	defer close(done)
	svr.printConnErr = false
	pingStream := client.Subscribe(PingPongTopic)
	poolDataStream := client.Subscribe(PoolDataTopic)
	client.Connect()
	var x int
	for n := 0; n < b.N; n++ {
		client.Send("PING 0")

		for x = 0; x < 2; x++ {
			select {
			case <-pingStream:
			case <-poolDataStream:
			case <-time.After(100 * time.Millisecond):
				fmt.Println("Failed (timeout)")
				b.Error("Failed (timeout)")
			}
		}
	}
}

func BenchmarkSendParallel(b *testing.B) {
	client, svr, done := getClientSvr(b)
	defer close(done)
	svr.printConnErr = false
	pingStream := client.Subscribe(PingPongTopic)
	poolDataStream := client.Subscribe(PoolDataTopic)
	client.Connect()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.Send("PING 0")

			for x := 0; x < 2; x++ {
				select {
				case <-pingStream:
				case <-poolDataStream:
				case <-time.After(100 * time.Millisecond):
					fmt.Println("Failed (timeout)")
					b.Error("Failed (timeout)")
				}
			}
		}
	})
}

// func BenchmarkReconnect(b *testing.B) {
// 	var (
// 		pingStream, poolDataStream chan interface{}
// 		oldConnect                 chan struct{}
// 	)
// 	oldPing := PingInterval
// 	PingInterval = 3000000 * time.Second
// 	defer func() { PingInterval = oldPing }()
// 	b.Logf("b.N is: %d\n", b.N)
// 	var x int
// 	var err error
// 	for n := 0; n < b.N; n++ {
// 		func() {
// 			client, svr, done := getClientSvr(b)
// 			defer close(done)
// 			svr.printConnErr = false
// 			err = client.Connect()
// 			if err != nil {
// 				b.Fatalf("Failed on connect: %s\n", err.Error())
// 			}
// 			pingStream = client.Subscribe(PingPongTopic)
// 			poolDataStream = client.Subscribe(PoolDataTopic)

// 			client.Send("PING 2")

// 			for x = 0; x < 2; x++ {
// 				select {
// 				case <-pingStream:
// 					// fmt.Println("in pingstream")
// 				case <-poolDataStream:
// 					// fmt.Println("in pooldatastream")
// 				case <-time.After(500 * time.Millisecond):
// 					// fmt.Println("Failed 1 (timeout)")
// 					b.Fatal("Failed 1 (timeout)")
// 				}
// 			}

// 			client.Unsubscribe(pingStream)
// 			client.Unsubscribe(poolDataStream)

// 			oldConnect = client.Connected()
// 			svr.conn.Close()

// 			for oldConnect == client.Connected() {
// 				// b.Log("Still the old chan")
// 				time.Sleep(100 * time.Microsecond)
// 			}

// 			select {
// 			case <-client.Joined():
// 			case <-time.After(500 * time.Millisecond):
// 				b.Fatal("Timed out waiting for reconnect")
// 			}

// 			pingStream = client.Subscribe(PingPongTopic)
// 			poolDataStream = client.Subscribe(PoolDataTopic)

// 			client.Send("PING 3")

// 			for x = 0; x < 2; x++ {
// 				select {
// 				case <-pingStream:
// 					// fmt.Println("in pingstream")
// 				case <-poolDataStream:
// 					// fmt.Println("in pooldatastream")
// 				case <-time.After(500 * time.Millisecond):
// 					// fmt.Println("Failed 2 (timeout)")
// 					b.Fatal("Failed 2 (timeout)")
// 				}
// 			}
// 			client.Unsubscribe(pingStream)
// 			client.Unsubscribe(poolDataStream)
// 			// time.Sleep(20 * time.Millisecond)
// 		}()
// 	}
// }
