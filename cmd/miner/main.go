package main

import (
	"fmt"

	"github.com/leviable/noso-go/internal/miner"
	"github.com/leviable/noso-go/internal/tcplib"
)

const (
	minerVer = "go-miner-0-1-0"
	// minerVer = "1.65"
)

func main() {
	var (
		resp string
		// err           error
		// currentBlock  int
		// targetBlock   int
		// currentStep   int
		// minerHashRate int
	)
	opts := miner.GetOpts()
	client := tcplib.NewTcpClient(opts)
	fmt.Printf("Client: %+v\n", client)

	client.SendChan <- fmt.Sprintf("JOIN %s", minerVer)

	comms := miner.Comms{
		CurrentBlock: make(chan int, 0),
		TargetBlock:  make(chan int, 0),
		CurrentStep:  make(chan int, 0),
		Hashes:       make(chan int, 0),
	}

	for {
		select {
		// case currentBlock <- currentBlockChan:
		// case targetBlock <- targetBlockChan:
		// case currentStep <- currentStepChan:
		// case minerHashRate <- hashRateChan:
		case resp = <-client.RecvChan:
			miner.Parse(comms, resp)
		}
	}
}
