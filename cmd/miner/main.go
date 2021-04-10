package main

import (
	"fmt"

	"github.com/leviable/noso-go/internal/opts"
	"github.com/leviable/noso-go/internal/tcplib"
)

const (
	minerVer = "go-miner-0-1-0"
	// minerVer = "1.65"
)

type Comms struct {
	currentBlock chan int
	targetBlock  chan int
	currentStep  chan int
	hashes       chan int
}

func main() {
	var (
		resp string
		// err           error
		// currentBlock  int
		// targetBlock   int
		// currentStep   int
		// minerHashRate int
	)
	opts := opts.GetOpts()
	client := tcplib.NewTcpClient(opts)
	fmt.Printf("Client: %+v\n", client)

	client.SendChan <- fmt.Sprintf("JOIN %s", minerVer)

	comms := Comms{
		currentBlock: make(chan int, 0),
		targetBlock:  make(chan int, 0),
		currentStep:  make(chan int, 0),
		hashes:       make(chan int, 0),
	}

	for {
		select {
		// case currentBlock <- currentBlockChan:
		// case targetBlock <- targetBlockChan:
		// case currentStep <- currentStepChan:
		// case minerHashRate <- hashRateChan:
		case resp = <-client.RecvChan:
			parse(comms, resp)
		}
	}
}

func parse(comms Comms, resp string) {
}
