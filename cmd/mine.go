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
	"github.com/leviable/noso-go/internal/miner"
	"github.com/spf13/cobra"
)

var (
	mineOpts = &miner.Opts{}
)

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "CPU mine for Noso coin",
	Long: `Connect to a specific Noso pool and CPU mine for Noso coin
Example usage:
./noso-go mine \
	--address noso.dukedog.io \
	--port 8082 \
	--password duke \
	--wallet Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF \
	--cpu 4
`,
	Run: func(cmd *cobra.Command, args []string) {
		miner.Mine(mineOpts)
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