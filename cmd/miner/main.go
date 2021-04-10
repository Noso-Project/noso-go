package main

import (
	"fmt"

	"github.com/leviable/noso-go/internal/miner"
)

const (
	minerVer = "go-miner-0-1-0"
)

func main() {
	var (
		// err          error
		resp         string
		poolAddr     string
		minerSeed    string
		targetBlock  int
		targetString string
		targetChars  int
		currentStep  int
		totalHashes  int
	)
	opts := miner.GetOpts()
	client := miner.NewTcpClient(opts)
	fmt.Printf("Client: %+v\n", client)

	client.SendChan <- fmt.Sprintf("JOIN %s", minerVer)

	comms := miner.NewComms()

	for {
		select {
		case poolAddr = <-comms.PoolAddr:
			fmt.Printf("PoolAddress is %s\n", poolAddr)
		case minerSeed = <-comms.MinerSeed:
			fmt.Printf("minerSeed is %s\n", minerSeed)
		case targetBlock = <-comms.TargetBlock:
			fmt.Printf("Target block is %d\n", targetBlock)
		case targetString = <-comms.TargetString:
			fmt.Printf("Target string is %s\n", targetString)
		case targetChars = <-comms.TargetChars:
			fmt.Printf("Target chars are %d\n", targetChars)
		case currentStep = <-comms.CurrentStep:
			fmt.Printf("Current step is %d\n", currentStep)
		case hashes := <-comms.Hashes:
			totalHashes += hashes
			fmt.Printf("Current hashes is %d\n", totalHashes)
		case resp = <-client.RecvChan:
			go miner.Parse(comms, resp)
		}
	}
}
