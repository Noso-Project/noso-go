package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	DUMMYADDR = "fakeurl.com"
	DUMMYPORT = 12345
)

func GetFixtures(t testing.TB) (*Client, *TcpServer, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	r := make(respMap)
	svr := NewTcpServer(ctx, t, r)
	client := NewClient(ctx, svr.Host, svr.Port)

	return client, svr, ctx, cancel
}

// TODO: Remove this if we no longer need it
// func TestMain(m *testing.M) {
// 	// logWriter = os.Stdout
// 	os.Exit(m.Run())
// }

func TestClientConnect(t *testing.T) {
	t.Run("new client", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		got := NewClient(ctx, DUMMYADDR, DUMMYPORT).poolAddr
		want := fmt.Sprintf("%s:%d", DUMMYADDR, DUMMYPORT)

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("connect refused", func(t *testing.T) {
		client, svr, _, cancel := GetFixtures(t)
		defer cancel()
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

		client, svr, _, cancel := GetFixtures(t)
		defer cancel()
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
		client, _, _, cancel := GetFixtures(t)
		defer cancel()

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
		client, _, _, cancel := GetFixtures(t)
		defer cancel()

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
	t.Run("join with bad password", func(t *testing.T) {
		client, svr, _, cancel := GetFixtures(t)
		defer cancel()

		svr.rMap[JOIN] = []string{PASSFAILED_default}

		joinStream, err := client.Subscribe(JoinTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(joinStream)
		// TODO: This might cause a hang in the client in a real world
		//       scenario. Investigate and improve if needed
		// broker publish will hang here if connect is not in it's own
		// goroutine
		go client.Connect()

		select {
		case got := <-joinStream:
			switch got.(type) {
			case PassFailed:
			default:
				t.Errorf("got %v, want passFailed", got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for passFailed resp from server")
		}
	})
	t.Run("reconnect on closed connection", func(t *testing.T) {
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		oldWait := ReconnectWait
		ReconnectWait = time.Microsecond
		defer func() { ReconnectWait = oldWait }()

		client, svr, _, cancel := GetFixtures(t)
		defer cancel()

		svr.printConnErr = false

		pingStream, err := client.Subscribe(PingPongTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
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
		// TODO: Should probably use a sync.Cond broadcast and/or a
		//       broker subscription to watch for a connected event

		oldConnectedChan := client.Connected()

		svr.conn.Close()

		for oldConnectedChan == client.Connected() {
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

		pingStream, err = client.Subscribe(PingPongTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(pingStream)

		select {
		case <-pingStream:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for server pong")
		}
	})
	t.Run("reconnect on read deadline exceeded", func(t *testing.T) {
		oldDeadline := DeadlineExceededTimeout
		DeadlineExceededTimeout = 10 * time.Millisecond
		defer func() { DeadlineExceededTimeout = oldDeadline }()

		oldWait := ReconnectWait
		ReconnectWait = time.Microsecond
		defer func() { ReconnectWait = oldWait }()

		client, svr, _, cancel := GetFixtures(t)
		defer cancel()

		svr.printConnErr = false

		pingStream, err := client.Subscribe(PingPongTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		client.Connect()

		// Get one pong, wait for deadline exceeded, then
		// wait for one more pong
		client.Send("PING 1")

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

		for oldConnectedChan == client.Connected() {
			time.Sleep(1 * time.Microsecond)
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

		pingStream, err = client.Subscribe(PingPongTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(pingStream)
		client.Send("PING 2")

		select {
		case <-pingStream:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for server pong")
		}
	})
	t.Run("reconnect on ALREADYCONNECTED", func(t *testing.T) {
		oldWait := ReconnectWait
		ReconnectWait = time.Microsecond
		defer func() { ReconnectWait = oldWait }()

		client, svr, ctx, cancel := GetFixtures(t)
		defer cancel()

		svr.rMap[JOIN] = []string{ALREADYCONNECTED_default, JOINOK_default}

		svr.printConnErr = false

		joinStream, err := client.Subscribe(JoinTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		joinPassthroughStream := make(chan interface{}, 0)

		// TODO: This is getting ugly, need to rethink connect and join
		//     : Probably need to internally spearate connect and join
		go func() {
			defer close(joinPassthroughStream)
			select {
			case <-ctx.Done():
				return
			case joinPassthroughStream <- <-joinStream:
			}
		}()

		oldConnectedChan := client.Connected()
		client.Connect()

		// Get ALREADYCONNECTED to force conn close and reconnect

		select {
		case msg := <-joinPassthroughStream:
			switch msg.(type) {
			case AlreadyConnected:
			default:
				t.Fatal("Expected ALREADYCONNECTED, got: ", msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for ALREADYCONNECTED")
		}

		// TODO: Should I be waiting for these streams to actually close?
		client.Unsubscribe(joinStream)

		// Subtle race condition here:
		//  - After closing the connection, its possible that
		//  - the Connected() and Joined() checks will work
		//  - because they are the old, already closed channels,
		//  - and then the Subscribe call goes to the old broker
		//
		// Need to wait for the client to re-init before checking
		// its connected/joined status

		start := time.Now()
		for oldConnectedChan == client.Connected() {
			// TODO: Set this back to 100 microseconds
			time.Sleep(100 * time.Microsecond)

			if time.Since(start) > 100*time.Millisecond {
				t.Fatal("Timed out waiting to reconnect")
			}
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
			t.Fatal("Timed out waiting to join pool")
		}
	})
}

func TestClientMessaging(t *testing.T) {
	t.Run("ping and pong", func(t *testing.T) {
		oldPing := PingInterval
		PingInterval = 10 * time.Millisecond
		defer func() { PingInterval = oldPing }()

		client, _, _, cancel := GetFixtures(t)
		defer cancel()

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		pongStream, err := client.Subscribe(PingPongTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(pongStream)

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

		client, svr, _, cancel := GetFixtures(t)
		defer cancel()

		svr.rMap[PING] = []string{POOLSTEPS_default}
		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		poolStepsStream, err := client.Subscribe(PoolStepsTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(poolStepsStream)

		select {
		case <-poolStepsStream:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for server poolsteps")
		}
	})
	t.Run("pooldata from joinOk", func(t *testing.T) {
		client, _, _, cancel := GetFixtures(t)
		defer cancel()

		poolDataStream, err := client.Subscribe(PoolDataTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(poolDataStream)

		go func() {
			err := client.Connect()
			if err != nil {
				t.Fatal("Got an error and didn't expect one: ", err)
			}
		}()

	loop:
		for {
			select {
			case msg := <-poolDataStream:
				switch msg.(type) {
				case JoinOk:
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

		client, _, _, cancel := GetFixtures(t)
		defer cancel()

		poolDataStream, err := client.Subscribe(PoolDataTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(poolDataStream)
		go func() {
			err := client.Connect()
			if err != nil {
				t.Fatal("Got an error and didn't expect one: ", err)
			}
		}()

		after := time.After(100 * time.Millisecond)
	loop:
		for {
			select {
			case msg := <-poolDataStream:
				switch msg.(type) {
				case Pong:
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

		client, svr, _, cancel := GetFixtures(t)
		defer cancel()

		svr.rMap[PING] = []string{POOLSTEPS_default}

		poolDataStream, err := client.Subscribe(PoolDataTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(poolDataStream)
		go func() {
			err := client.Connect()
			if err != nil {
				t.Fatal("Got an error and didn't expect one: ", err)
			}
		}()

		after := time.After(100 * time.Millisecond)
	loop:
		for {
			select {
			case msg := <-poolDataStream:
				switch msg.(type) {
				case PoolSteps:
					break loop
				default:
					continue
				}
			case <-after:
				t.Fatal("Timed out waiting for pooldata in poolSteps msg")
			}
		}
	})
	t.Run("step and stepOk", func(t *testing.T) {
		client, _, _, cancel := GetFixtures(t)
		defer cancel()

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		stepOkStream, err := client.Subscribe(StepOkTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(stepOkStream)

		client.Send("STEP 35360 0e0000cyk 8b9554VKg 9")

		select {
		case resp := <-stepOkStream:
			switch resp.(type) {
			case StepOk:
				got := resp.(StepOk).PopValue
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
	t.Run("publish solution", func(t *testing.T) {
		client, _, _, cancel := GetFixtures(t)
		defer cancel()

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		solStream, err := client.Subscribe(SolutionTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(solStream)

		client.Publish(Solution{
			Block:     12345,
			Seed:      "seedstring",
			HashStr:   "hashstring",
			TargetLen: 54321,
		})

		select {
		case sol := <-solStream:
			switch sol.(type) {
			case Solution:
				got := sol.(Solution).Block
				want := 12345

				if got != want {
					t.Errorf("got %d, want %d", got, want)
				}
			default:
				t.Errorf("Expected Solution msg, but got %v", sol)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for server StepOk")
		}
	})
	t.Run("publish job", func(t *testing.T) {
		client, _, _, cancel := GetFixtures(t)
		defer cancel()

		err := client.Connect()
		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err)
		}

		jobStream, err := client.Subscribe(JobTopic)
		if err != nil {
			t.Fatal("Got an error and didn't expect one:", err)
		}
		defer client.Unsubscribe(jobStream)

		publishedJob := Job{
			PoolAddr:      "pooladdr",
			MinerSeedBase: "seedminer",
			MinerSeed:     "seedpostfix",
			TargetString:  "targetstring",
			TargetChars:   11,
			Diff:          111,
			Block:         12345,
			Step:          2,
			PoolDepth:     3,
		}
		client.Publish(publishedJob)

		select {
		case job := <-jobStream:
			switch job.(type) {
			case Job:
				got := job
				want := publishedJob

				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			default:
				t.Errorf("Expected Job msg, but got %v", job)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for Job")
		}
	})
}

func BenchmarkSend(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)
	client, svr, _, cancel := GetFixtures(b)
	defer cancel()

	svr.printConnErr = false
	pingStream, err := client.Subscribe(PingPongTopic)
	if err != nil {
		b.Fatal("Got an error and didn't expect one:", err)
	}
	client.Connect()
	poolDataStream, err := client.Subscribe(PoolDataTopic)
	if err != nil {
		b.Fatal("Got an error and didn't expect one:", err)
	}
	var x int
	for n := 0; n < b.N; n++ {
		client.Send("PING 0")

		for x = 0; x < 2; x++ {
			select {
			case <-pingStream:
			case <-poolDataStream:
			case <-time.After(100 * time.Millisecond):
				b.Log("Failed (timeout)")
				b.Error("Failed (timeout)")
			}
		}
	}
}

func BenchmarkSendParallel(b *testing.B) {
	client, svr, _, cancel := GetFixtures(b)
	defer cancel()
	svr.printConnErr = false
	pingStream, err := client.Subscribe(PingPongTopic)
	if err != nil {
		b.Fatal("Got an error and didn't expect one:", err)
	}
	client.Connect()
	poolDataStream, err := client.Subscribe(PoolDataTopic)
	if err != nil {
		b.Fatal("Got an error and didn't expect one:", err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.Send("PING 0")

			for x := 0; x < 2; x++ {
				select {
				case <-pingStream:
				case <-poolDataStream:
				case <-time.After(100 * time.Millisecond):
					b.Log("Failed (timeout)")
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
// 			client, svr, done := GetFixtures(b)
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
// 					// b.Log("in pingstream")
// 				case <-poolDataStream:
// 					// b.Log("in pooldatastream")
// 				case <-time.After(500 * time.Millisecond):
// 					// b.Log("Failed 1 (timeout)")
// 					b.Fatal("Failed 1 (timeout)")
// 				}
// 			}

// 			client.Unsubscribe(pingStream)
// 			client.Unsubscribe(poolDataStream)

// 			oldConnect = client.Connected()
// 			svr.conn.Close()

// 			for oldConnect == client.Connected() {
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
// 					// b.Log("in pingstream")
// 				case <-poolDataStream:
// 					// b.Log("in pooldatastream")
// 				case <-time.After(500 * time.Millisecond):
// 					// b.Log("Failed 2 (timeout)")
// 					b.Fatal("Failed 2 (timeout)")
// 				}
// 			}
// 			client.Unsubscribe(pingStream)
// 			client.Unsubscribe(poolDataStream)
// 			// time.Sleep(20 * time.Millisecond)
// 		}()
// 	}
// }
