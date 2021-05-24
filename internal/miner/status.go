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
	"os"
	"strconv"
	"strings"
	"time"
)

func GetPoolStatus(opts *Opts) {
	var (
		resp string
	)

	fmt.Printf("Connecting to %s:%d with password %s\n", opts.IpAddr, opts.IpPort, opts.PoolPw)
	fmt.Printf("Using wallet address: %s\n", opts.Wallet)
	comms := NewComms()
	client := NewTcpClient(opts, comms, false)

	client.SendChan <- "STATUS"

loop:
	for {
		select {
		case resp = <-client.RecvChan:
			go Parse(comms, opts.IpAddr, opts.Wallet, 0, resp)
		case status := <-comms.PoolStatus:
			fmt.Printf("Got status from pool: %+v\n", status)
			break loop
		case <-time.After(5 * time.Second):
			fmt.Println("Failed to get a response back from the pool")
			os.Exit(1)
		}
	}
}

type PoolStatus struct {
	HashRateRaw string
	HashRate    string
	FeeRaw      int
	Fee         string
	ShareRaw    int
	Share       string
	MinerCnt    string
	Miners      []MinerInfo
}

// s[] -> STATUS {hashrate} {fee} {share} {minerCount} [list of miners: {address}:{balance}:{blocks_until_paymet}]
func NewPoolStatus(s []string) PoolStatus {
	var (
		fee   string
		share string
	)

	hrRaw := s[0]
	hr := formatHashRate(hrRaw + "000")
	feeRaw, err := strconv.Atoi(s[1])
	if err != nil {
		fmt.Println("Error converting Fee to int")
	} else {
		fee = fmt.Sprintf("%.2f%%", float64(feeRaw)/100)
	}
	shareRaw, err := strconv.Atoi(s[2])
	if err != nil {
		fmt.Println("Error converting Share to int")
	} else {
		share = fmt.Sprintf("%.2f%%", float64(shareRaw)/100)
	}
	minerCnt := s[3]

	miners := formatMiners(s[4:])

	return PoolStatus{
		HashRateRaw: hrRaw,
		HashRate:    hr,
		FeeRaw:      feeRaw,
		Fee:         fee,
		ShareRaw:    shareRaw,
		Share:       share,
		MinerCnt:    minerCnt,
		Miners:      miners,
	}
}

type MinerInfo struct {
	Address           string
	Balance           string
	BalanceHR         string
	BlocksTillPayment string
}

func newMinerInfo(m string) MinerInfo {
	split := strings.Split(m, ":")
	address := split[0]
	balance := split[1]
	btp := split[2]
	return MinerInfo{
		Address:           address,
		Balance:           balance,
		BalanceHR:         formatBalance(balance),
		BlocksTillPayment: btp,
	}
}

func formatMiners(m []string) []MinerInfo {
	miners := make([]MinerInfo, 0)

	for _, miner := range m {
		if len(miner) == 0 {
			continue
		}
		miners = append(miners, newMinerInfo(miner))
	}

	return miners
}
