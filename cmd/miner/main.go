package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"net/http"
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
		currentDiff  int
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
	jobComms := miner.NewJobComms()
	go miner.JobFeeder(comms, jobComms)

	// Start the Solutions Manager goroutine
	solComms := miner.NewSolutionComms(client.SendChan)
	go miner.SolutionManager(solComms)

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

	// TODO: Sending individual info (block, chars, string, etc
	//       will probably lead to a race condition. Send a
	//       BlockUpdate struct instead with all info?
	for {
		select {
		case poolAddr = <-comms.PoolAddr:
			jobComms.PoolAddr <- poolAddr
		case minerSeed = <-comms.MinerSeed:
			jobComms.MinerSeed <- minerSeed
		case targetString = <-comms.TargetString:
			jobComms.TargetString <- targetString
		case targetChars = <-comms.TargetChars:
			jobComms.TargetChars <- targetChars
		case targetBlock = <-comms.Block:
			jobComms.Block <- targetBlock
			solComms.Block <- targetBlock
		case currentStep = <-comms.Step:
			jobComms.Step <- currentStep
			solComms.Step <- currentStep
		case currentDiff = <-comms.Diff:
			jobComms.Diff <- currentDiff
			solComms.Diff <- currentDiff
		case <-comms.StepSolved:
			stepsSolved += 1
			fmt.Printf("Miner has solved %d steps\n", stepsSolved)
		case sol := <-comms.Solutions:
			solComms.Solution <- sol
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
