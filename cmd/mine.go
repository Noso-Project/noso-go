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
	"os"

	"github.com/Noso-Project/noso-go/internal/miner"
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
		if mineOpts.Cpu < 1 {
			cmd.PrintErrln("Error: --cpu cannot be less than 1")
			os.Exit(1)
		}

		ipAddr, err := lookupIP(mineOpts.IpAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get IP address for domain: %v\n", err)
			os.Exit(1)
		}
		mineOpts.IpAddr = ipAddr

		miner.Mine(mineOpts)
	},
}

func init() {
	rootCmd.AddCommand(mineCmd)

	mineCmd.Flags().StringVarP(&mineOpts.IpAddr, "address", "a", "", "Pool IP address (e.g. 'noso.dukedog.io' or '75.45.193.238'")
	mineCmd.Flags().IntVar(&mineOpts.IpPort, "port", 8082, "Pool port")
	mineCmd.Flags().StringVarP(&mineOpts.PoolPw, "password", "p", "", "Pool password")
	mineCmd.Flags().StringVarP(&mineOpts.Wallet, "wallet", "w", "", "Noso wallet address to send payments to")
	mineCmd.Flags().IntVarP(&mineOpts.Cpu, "cpu", "c", 4, "Number of CPU cores to use")
	mineCmd.Flags().BoolVarP(&mineOpts.ShowPop, "show-pop", "", false, "Show PoP solutions in output")
	mineCmd.Flags().IntVar(&mineOpts.StatusInterval, "status-interval", 60, "Status Interval Timer (in seconds)")

	mineCmd.MarkFlagRequired("address")
	mineCmd.MarkFlagRequired("password")
	mineCmd.MarkFlagRequired("wallet")

	mineCmd.Flags().SortFlags = false
	mineCmd.Flags().PrintDefaults()
}
