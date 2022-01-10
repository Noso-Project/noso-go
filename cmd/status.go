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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Returns status of Noso pool",
	Long: `Returns the status (miners connected, hashrate, etc) of a given Noso pool. For example:
Example usage:

List available pools
./noso-go status --list

List info about a specific pool
./noso-go status devnoso --info

Get status of a pool
./noso-go status devnoso    --wallet <your wallet address>
./noso-go status dukedog.io --wallet <your wallet address>
./noso-go status mining.moe --wallet <your wallet address>
./noso-go status russiapool --wallet <your wallet address>
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if list {
			return nil
		}
		if len(args) < 1 {
			return errors.New("requires a pool name (e.g. 'noso-go status devnoso')")
		}
		poolName := strings.ToLower(args[0])
		if _, ok := pools[poolName]; !ok {
			errMsg := fmt.Sprintf("Unrecognized pool name %q. Use 'noso-go status --list' for list of pools", poolName)
			return errors.New(errMsg)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if list {
			listPools()
			return
		}

		poolName := strings.ToLower(args[0])
		pool := pools[poolName]

		if info {
			printPoolInfo(pool)
			return
		}

		if len(poolOpts.Wallets) == 0 {
			cmd.PrintErrln("Error: required flag(s) \"--wallet\" not set")
			cmd.PrintErrf("Run '%v --help' for usage.\n", cmd.CommandPath())
			os.Exit(1)
		}

		poolOpts.IpAddr = pool.opts.IpAddr
		poolOpts.IpPort = pool.opts.IpPort
		poolOpts.PoolPw = pool.opts.PoolPw

		// TODO: add this back in
		// miner.GetPoolStatus(poolOpts)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&list, "list", "l", false, "List known pool names")
	statusCmd.Flags().BoolVarP(&info, "info", "i", false, "Print Pool information and exit")
	statusCmd.Flags().StringSliceVarP(&poolOpts.Wallets, "wallet", "w", []string{}, "Noso wallet address to send payments to")

	statusCmd.Flags().SortFlags = false
	statusCmd.Flags().PrintDefaults()
}
