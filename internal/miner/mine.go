/*
Copyright © 2021 Levi Noecker <levi.noecker@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package miner

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	HEADER = `
# ##########################################################
#
# noso-go %s by levi.noecker@gmail.com (c)2021
# https://github.com/Noso-Project/noso-go
# Commit: %s
#
# ##########################################################
`
)

func Mine(opts *Opts) {
	var (

		// last response from pool
		resp string

		// state vars
		workerReports     map[string]Report
		poolAddr          string
		minerSeed         string
		targetBlock       int
		targetString      string
		targetChars       int
		currentStep       int
		currentDiff       int
		poolDepth         int
		stepsSent         int
		stepsAccepted     int
		sharesEarned      int
		sharesEarnedBlk   int
		blocksTillPayment int
		balance           string
		paymentRequested  time.Time

		// hash rate info
		totalHashes  int
		hashRate     int
		poolHashRate string

		// syncing
		m sync.RWMutex
	)
	fmt.Printf(HEADER, Version, Commit)

	workerReports = make(map[string]Report)

	// Set a date in the past so we can request payment immediately if we
	// have a vested balance
	balance = "0"
	paymentRequested = time.Now().Add(-3 * time.Hour)

	fmt.Printf("Connecting to %s:%d with password %s\n", opts.IpAddr, opts.IpPort, opts.PoolPw)
	fmt.Printf("Using wallet address: %s\n", opts.Wallet)
	fmt.Printf("Number of CPU cores to use: %d\n", opts.Cpu)
	comms := NewComms()
	client := NewTcpClient(opts, comms, true, true)

	// Start the job feeder goroutine
	jobComms := NewJobComms()
	go JobFeeder(comms, jobComms)

	// Start the Solutions Manager goroutine
	solComms := NewSolutionComms(client.SendChan)
	go SolutionManager(solComms, opts.ShowPop)

	// Start the miner goroutines
	ready := make(chan bool, 0)
	for x := 1; x <= opts.Cpu; x++ {
		go Miner(strconv.Itoa(x), comms, ready)
	}

	// TODO: Need to do a sync broadcast for ready
	go func() {

		for running := true; running; {
			m.RLock()
			if targetChars == 0 || targetBlock == 0 || targetString == "" {
				time.Sleep(100 * time.Millisecond)
			} else {
				running = false
			}
			m.RUnlock()
		}
		close(ready)
	}()

	// Create the payments.csv file if it doesn't already exist
	CreateLogPaymentsFile()

	// Print a reward status every 60 seconds
	go func() {
		for {
			select {
			case <-time.After(60 * time.Second):
				m.RLock()
				hr := hashRate
				bal := balance
				m.RUnlock()
				fmt.Printf(
					statusMsg,
					targetBlock,
					formatHashRate(strconv.Itoa(hr)),
					formatHashRate(poolHashRate),
					formatBalance(bal),
					blocksTillPayment,
					stepsSent,
					stepsAccepted,
				)
			}
		}
	}()

	// TODO: Sending individual info (block, chars, string, etc
	//       will probably lead to a race condition. Send a
	//       BlockUpdate struct instead with all info?
	for {
		select {
		case poolAddr = <-comms.PoolAddr:
			jobComms.PoolAddr <- poolAddr
		case minerSeed = <-comms.MinerSeed:
			jobComms.MinerSeed <- minerSeed
		case ts := <-comms.TargetString:
			m.Lock()
			targetString = ts
			m.Unlock()
			jobComms.TargetString <- ts
		case tc := <-comms.TargetChars:
			m.Lock()
			targetChars = tc
			m.Unlock()
			jobComms.TargetChars <- tc
		case newBlock := <-comms.Block:
			m.Lock()
			targetBlock = newBlock
			m.Unlock()
			jobComms.Block <- newBlock
			solComms.Block <- newBlock
		case currentStep = <-comms.Step:
			jobComms.Step <- currentStep
			solComms.Step <- currentStep
		case currentDiff = <-comms.Diff:
			jobComms.Diff <- currentDiff
			solComms.Diff <- currentDiff
		case poolDepth = <-comms.PoolDepth:
			jobComms.PoolDepth <- poolDepth
		case bal := <-comms.Balance:
			m.Lock()
			balance = bal
			m.Unlock()
		case poolHashRate = <-comms.PoolHashRate:
		case blocksTillPayment = <-comms.BlocksTillPayment:
			// If we have a non-zero balance
			// And our balance is fully vested
			// And we haven't requested payment in at least 10 minutes
			if balance != "0" && blocksTillPayment > 0 && time.Since(paymentRequested) > 10*time.Minute {
				client.SendChan <- "PAYMENT"
				LogPaymentReq(opts.IpAddr, opts.Wallet, targetBlock, balance)
				paymentRequested = time.Now()
			}
		case <-solComms.StepSent:
			stepsSent++
		case shares := <-comms.StepSolved:
			stepsAccepted++
			sharesEarned += shares
			sharesEarnedBlk += shares
		case <-comms.StepFailed:
		case sol := <-comms.Solutions:
			solComms.Solution <- sol
		case report := <-comms.Reports:
			// TODO: do rolling average instead of all time
			workerReports[report.WorkerNum] = report

			hr := 0
			for _, rep := range workerReports {
				dur := float64(rep.Duration) / float64(time.Second)
				hr += int(float64(rep.Hashes) / dur)
			}
			totalHashes += report.Hashes
			m.Lock()
			hashRate = hr
			m.Unlock()
			comms.HashRate <- hr
		case resp = <-client.RecvChan:
			go Parse(comms, opts.IpAddr, opts.Wallet, targetBlock, resp)
		}
	}
}

var hashNameMap = map[int]string{
	0:  "Hashes",
	1:  "Hashes",
	2:  "Hashes",
	3:  "Kilohashes",
	4:  "Kilohashes",
	5:  "Kilohashes",
	6:  "Megahashes",
	7:  "Megahashes",
	8:  "Megahashes",
	9:  "Gigahashes",
	10: "Gigahashes",
	11: "Gigahashes",
	12: "Terahashes",
	13: "Terahashes",
	14: "Terahashes",
	15: "Petahashes",
	16: "Petahashes",
	17: "Petahashes",
	18: "Exahashes",
	19: "Exahashes",
	20: "Exahashes",
	21: "Zettahashes",
	22: "Zettahashes",
	23: "Zettahashes",
}

// func formatHashRate(hashRate int) string {
// 	return strconv.Itoa(hashRate)
// }

func formatBalance(balance string) string {
	return fmt.Sprintf("%s Noso", parseAmount(balance))
}

const statusMsg = `
************************************

Miner Status

Current Block       : %d

Miner Hash Rate     : %s
Pool Hash Rate      : %s

Pool Balance        : %s
Blocks Till Payment : %d

Proof of Participation
----------------------
PoP Sent            : %d
PoP Accepted        : %d

************************************

`
