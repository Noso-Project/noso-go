package main

import (
	"fmt"

	"github.com/leviable/noso-go/internal/opts"
	"github.com/leviable/noso-go/internal/tcplib"
)

const (
	// minerVer = "go-miner-0-1-0"
	minerVer = "1.65"
)

func main() {
	var (
		resp_raw string
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

	for {
		select {
		// case currentBlock <- currentBlockChan:
		// case targetBlock <- targetBlockChan:
		// case currentStep <- currentStepChan:
		// case minerHashRate <- hashRateChan:
		case resp_raw = <-client.RecvChan:
			fmt.Println(resp_raw)
		}
	}
}
