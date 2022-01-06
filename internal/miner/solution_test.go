package miner

import (
	"bufio"
	"context"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/Noso-Project/noso-go/internal/common"
	"github.com/fortytw2/leaktest"
)

func TestSolutionManager(t *testing.T) {
	if _, present := os.LookupEnv("LEAKTEST"); present {
		defer leaktest.Check(t)()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	conn, svr := net.Pipe()
	defer conn.Close()
	defer svr.Close()

	client := common.NewClientWithConn(ctx, conn)

	wg.Add(1)
	go SolutionManager(ctx, client, &wg)
	wg.Wait()

	sol := common.Solution{
		Block:     12345,
		Seed:      "seedstring",
		HashStr:   "hashstring",
		TargetLen: 54321,
	}

	client.Publish(ctx, sol)

	scanner := bufio.NewScanner(svr)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		t.Fatal("Got an error reading from server connection:", err)
	}

	// There is an extra space in front of the "got" value because the client auth string
	// normally preceeds it
	got := scanner.Text()
	want := " " + sol.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// func BenchmarkSolutionManager(b *testing.B) {
// 	b.Logf("b.N is: %d\n", b.N)

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	var wg sync.WaitGroup

// 	conn, svr := net.Pipe()
// 	defer conn.Close()
// 	defer svr.Close()

// 	client := common.NewClientWithConn(ctx, conn)

// 	wg.Add(1)
// 	go JobManager(ctx, client, &wg)
// 	wg.Wait()

// 	// A miner will publish a request to the JobTopic, requesting a jobStream,
// 	// and including in the request a channel to receive the jobStream on
// 	jobStream := requestJobStream(ctx, client)

// 	// Send a JOINOK from svr to client so PoolData info gets published
// 	// The JobManager will get the PoolData message, build a new Job,
// 	// and put it on the jobStream
// 	go func() {
// 		fmt.Fprintln(svr, JOINOK)
// 		pongTicker := time.NewTicker(11 * time.Millisecond)
// 		poolStepsTicker := time.NewTicker(17 * time.Millisecond)
// 		for count := 0; ; count++ {
// 			select {
// 			case <-pongTicker.C:
// 				fmt.Fprintln(svr, PONG)
// 			case <-poolStepsTicker.C:
// 				fmt.Fprintln(svr, fmt.Sprintf(POOLSTEPS, count, count%10))
// 			}
// 		}
// 	}()

// 	for n := 0; n < b.N; n++ {
// 		<-jobStream
// 	}
// }
