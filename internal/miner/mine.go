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
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		btpNote           string

		// hash rate info
		totalHashes  int
		hashRate     int
		poolHashRate string

		// syncing
		m sync.RWMutex
	)

	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)

	fileName := filepath.Join(exPath, "noso-go.log")
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error writing to log file: ", err)
	} else {
		defer file.Close()
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
		log.Printf("Writing logs to: %s", fileName)
	}
	log.Printf(HEADER, Version, Commit)

	workerReports = make(map[string]Report)

	// Set a date in the past so we can request payment immediately if we
	// have a vested balance
	balance = "0"
	paymentRequested = time.Now().Add(-3 * time.Hour)

	log.Printf("Connecting to %s:%d with password %s\n", opts.IpAddr, opts.IpPort, opts.PoolPw)
	log.Printf("Using wallet address(es): %s\n", strings.Join(opts.Wallets, " "))
	log.Printf("Number of CPU cores to use: %d\n", opts.Cpu)
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

	// Print a reward status every StatusInterval seconds (default 60)
	go func() {
		for {
			select {
			case <-time.After(time.Duration(opts.StatusInterval) * time.Second):
				m.RLock()
				hr := hashRate
				bal := balance
				m.RUnlock()
				log.Printf(
					statusMsg,
					opts.CurrentWallet,
					targetBlock,
					formatHashRate(strconv.Itoa(hr)),
					formatHashRate(poolHashRate),
					formatBalance(bal),
					blocksTillPayment,
					btpNote,
					stepsSent,
					stepsAccepted,
				)
			}
		}
	}()

	// TODO: Sending individual info (block, chars, string, etc
	//       will probably lead to a race condition. Send a
	//       BlockUpdate struct instead with all info?
main:
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
				LogPaymentReq(opts.IpAddr, opts.CurrentWallet, targetBlock, balance)
				paymentRequested = time.Now()
			} else if blocksTillPayment > 0 {
				btpNote = fmt.Sprint(`(* Note: A positive number here means you will
                            receive a payment as soon as the pool finds a block)`)
			} else {
				btpNote = ""
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
			go Parse(comms, opts.IpAddr, opts.CurrentWallet, targetBlock, resp)
		case <-comms.Disconnected:
			if opts.ExitOnRetry {
				break main
			}
		}
	}

	if opts.ExitOnRetry {
		// TODO: This will exit with code 0. Should it be non-zero?
		fmt.Println("Connection lost and --exit-on-retry flag is True. Exiting")
	}
}

const statusMsg = `
************************************

Miner Status

Miner's Wallet Addr : %s

Current Block       : %d

Miner Hash Rate     : %s
Pool Hash Rate      : %s

Pool Balance        : %s
Blocks Till Payment : %d %s


Proof of Participation
----------------------
PoP Sent            : %d
PoP Accepted        : %d

************************************

`
