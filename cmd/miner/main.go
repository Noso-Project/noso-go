package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "net/http/pprof"

	"github.com/leviable/noso-go/internal/miner"
)

const (
	minerVer = "noso-go-0-1-0"
)

func main() {
	var (
		// err          error

		// last response from pool
		resp string

		// state vars
		start        time.Time
		poolAddr     string
		minerSeed    string
		targetBlock  int
		targetString string
		targetChars  int
		currentStep  int
		stepsSolved  int

		// hash rate info
		totalHashes int
		hashRate    int
	)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	opts := miner.GetOpts()
	comms := miner.NewComms()
	client := miner.NewTcpClient(minerVer, opts, comms)

	// TODO: Need to handle join failures / fail-overs
	// client.SendChan <- fmt.Sprintf("JOIN %s", minerVer)

	// Start the job feeder goroutine
	go miner.JobFeeder(comms)

	// Start the miner goroutines
	ready := make(chan bool, 0)
	for x := 1; x <= opts.Cpu; x++ {
		go miner.Miner(strconv.Itoa(x), comms, ready)
	}

	// TODO: Need to do a sync broadcast for ready
	go func() {
		for targetChars == 0 || targetBlock == 0 || targetString == "" {
			time.Sleep(100 * time.Millisecond)
		}
		close(ready)
	}()

	start = time.Now()

	for {
		select {
		case poolAddr = <-comms.PoolAddr:
			comms.NewPoolAddr <- poolAddr
		case minerSeed = <-comms.MinerSeed:
			comms.NewMinerSeed <- minerSeed
		case targetBlock = <-comms.TargetBlock:
			comms.NewBlock <- targetBlock
		case targetString = <-comms.TargetString:
			comms.NewString <- targetString
		case targetChars = <-comms.TargetChars:
			comms.NewChars <- targetChars
		case currentStep = <-comms.CurrentStep:
			comms.NewStep <- currentStep
		case <-comms.StepSolved:
			stepsSolved += 1
			fmt.Printf("Miner has solved %d steps\n", stepsSolved)
		case report := <-comms.Reports:
			// TODO: do rolling average instead of all time
			totalHashes += report.Hashes
			timeSince := time.Since(start)
			dur := float64(timeSince) / float64(time.Second)
			hashRate = int(float64(totalHashes) / dur)
			comms.HashRate <- hashRate
		case resp = <-client.RecvChan:
			go miner.Parse(comms, resp)
		}
	}
}
