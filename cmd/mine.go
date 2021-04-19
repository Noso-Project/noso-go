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
package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/leviable/noso-go/internal/miner"
	"github.com/spf13/cobra"
)

var mineOpts = &miner.Opts{}

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "CPU mine for Noso coin",
	Long: `Connect to a Noso pool and CPU mine for Noso coin
Example usage:
./noso-go mine \
	--address noso.dukedog.io \
	--port 8082 \
	--password duke \
	--wallet Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF \
	--cpu 4
`,
	Run: func(cmd *cobra.Command, args []string) {
		mine(mineOpts)
	},
}

func init() {
	rootCmd.AddCommand(mineCmd)

	mineCmd.Flags().StringVarP(&mineOpts.IpAddr, "address", "a", "", "Pool IP address (e.g. 'noso.dukedog.io' or '75.45.193.238'")
	mineCmd.Flags().IntVar(&mineOpts.IpPort, "port", 8082, "Pool port")
	mineCmd.Flags().StringVarP(&mineOpts.PoolPw, "password", "p", "", "Pool password")
	mineCmd.Flags().StringVarP(&mineOpts.Wallet, "wallet", "w", "", "Noso wallet address to send payments to")
	mineCmd.Flags().IntVarP(&mineOpts.Cpu, "cpu", "c", 0, "Number of CPU cores to use")

	mineCmd.MarkFlagRequired("address")
	mineCmd.MarkFlagRequired("password")
	mineCmd.MarkFlagRequired("wallet")
}

func mine(opts *miner.Opts) {
	var (

		// last response from pool
		resp string

		// state vars
		workerReports     map[string]miner.Report
		poolAddr          string
		minerSeed         string
		targetBlock       int
		targetString      string
		targetChars       int
		currentStep       int
		currentDiff       int
		stepsSolved       int
		blocksTillPayment int
		balance           string
		paymentRequested  time.Time

		// hash rate info
		totalHashes int
		hashRate    int
	)

	workerReports = make(map[string]miner.Report)

	// Set a date in the past so we can request payment immediately if we
	// have a vested balance
	balance = "0"
	paymentRequested = time.Now().Add(-3 * time.Hour)

	comms := miner.NewComms()
	client := miner.NewTcpClient(opts, comms)

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

	// Create the payments.csv file if it doesn't already exist
	miner.CreateLogPaymentsFile()

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
		case balance = <-comms.Balance:
		case blocksTillPayment = <-comms.BlocksTillPayment:
			// If we have a non-zero balance
			// And our balance is fully vested
			// And we haven't requested payment in at least 10 minutes
			if balance != "0" && blocksTillPayment > 0 && time.Since(paymentRequested) > 10*time.Minute {
				client.SendChan <- "PAYMENT"
				miner.LogPaymentReq(opts.IpAddr, opts.Wallet, targetBlock)
				paymentRequested = time.Now()
			}
		case <-comms.StepSolved:
			stepsSolved += 1
			fmt.Printf("Miner has solved %d steps\n", stepsSolved)
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
			go miner.Parse(comms, opts.IpAddr, opts.Wallet, targetBlock, resp)
		}
	}
}
