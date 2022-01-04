package miner

import (
	"bufio"
	"context"
	"net"
	"sync"
	"testing"

	"github.com/Noso-Project/noso-go/internal/common"
)

func TestSolutionManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	conn, svr := net.Pipe()

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

	client.Publish(sol)

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
