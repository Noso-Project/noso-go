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
	"time"
)

const (
	HEADER = `# ##########################################################
#
# noso-go %s by levi.noecker@gmail.com (c)2021
# https://github.com/leviable/noso-go
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
		popSlice          []int
		popCount          int
		stepsAccepted     int
		sharesEarned      int
		sharesEarnedBlk   int
		blocksTillPayment int
		balance           string
		paymentRequested  time.Time

		// hash rate info
		totalHashes int
		hashRate    int
	)
	fmt.Printf(HEADER, Version)

	workerReports = make(map[string]Report)
	popSlice = make([]int, 0)

	// Set a date in the past so we can request payment immediately if we
	// have a vested balance
	balance = "0"
	paymentRequested = time.Now().Add(-3 * time.Hour)

	fmt.Printf("Connecting to %s:%d with password %s\n", opts.IpAddr, opts.IpPort, opts.PoolPw)
	fmt.Printf("Using wallet address: %s\n", opts.Wallet)
	fmt.Printf("Number of CPU cores to use: %d\n", opts.Cpu)
	comms := NewComms()
	client := NewTcpClient(opts, comms)

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
		for targetChars == 0 || targetBlock == 0 || targetString == "" {
			time.Sleep(100 * time.Millisecond)
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
				fmt.Printf(
					rewardMsg,
					targetBlock,     // For Block %d
					sharesEarnedBlk, // Shares Earned :
					len(popSlice),   // PoP Sent :
					sharesEarned,    // Shares Earned :
					popCount,        // PoP Sent :
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
		case targetString = <-comms.TargetString:
			jobComms.TargetString <- targetString
		case targetChars = <-comms.TargetChars:
			jobComms.TargetChars <- targetChars
		case newBlock := <-comms.Block:
			if newBlock != targetBlock {
				stepsAccepted += sumSteps(popSlice)
				popSlice = make([]int, 0)
				sharesEarned += sharesEarnedBlk
				sharesEarnedBlk = 0
			}
			targetBlock = newBlock
			jobComms.Block <- targetBlock
			solComms.Block <- targetBlock
		case currentStep = <-comms.Step:
			jobComms.Step <- currentStep
			solComms.Step <- currentStep
		case currentDiff = <-comms.Diff:
			jobComms.Diff <- currentDiff
			solComms.Diff <- currentDiff
		case poolDepth = <-comms.PoolDepth:
			jobComms.PoolDepth <- poolDepth
		case balance = <-comms.Balance:
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
			popSlice = append(popSlice, 0)
			popCount++
		case shares := <-comms.StepSolved:
			if len(popSlice) > 0 {
				popSlice[len(popSlice)-1]++
			}
			sharesEarned += shares
			sharesEarnedBlk += shares
		case <-comms.StepFailed:
			if len(popSlice) > 0 && popSlice[len(popSlice)-1] > 0 {
				popSlice[len(popSlice)-1]--
			}
		case sol := <-comms.Solutions:
			solComms.Solution <- sol
		case report := <-comms.Reports:
			// TODO: do rolling average instead of all time
			workerReports[report.WorkerNum] = report

			hashRate = 0
			for _, rep := range workerReports {
				dur := float64(rep.Duration) / float64(time.Second)
				hashRate += int(float64(rep.Hashes) / dur)
			}
			totalHashes += report.Hashes
			comms.HashRate <- hashRate
		case resp = <-client.RecvChan:
			go Parse(comms, opts.IpAddr, opts.Wallet, targetBlock, resp)
		}
	}
}

func sumSteps(popSlice []int) (sum int) {
	if len(popSlice) == 0 {
		return 0
	}
	for _, v := range popSlice {
		sum += v
	}

	return
}

const rewardMsg = `
************************************

Current Rewards Status

For Block %d
---------------
Shares Earned  : %d
PoP Sent       : %d

Total
---------------
Shares Earned  : %d
PoP Sent       : %d

************************************

`
