/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/Noso-Project/noso-go/internal/miner"
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
			return errors.New("Unrecognized pool name. Use 'noso-go status --list' for list of pools")
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
			printPoolInfo(poolName, pool)
			return
		}

		if poolOpts.Wallet == "" {
			cmd.PrintErrln("Error: required flag(s) \"--wallet\" not set")
			cmd.PrintErrf("Run '%v --help' for usage.\n", cmd.CommandPath())
			os.Exit(1)
		}

		poolOpts.IpAddr = pool.IpAddr
		poolOpts.IpPort = pool.IpPort
		poolOpts.PoolPw = pool.PoolPw

		miner.GetPoolStatus(poolOpts)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&list, "list", "l", false, "List known pool names")
	statusCmd.Flags().BoolVarP(&info, "info", "i", false, "Print Pool information and exit")
	statusCmd.Flags().StringVarP(&poolOpts.Wallet, "wallet", "w", "", "Noso wallet address to send payments to")

	statusCmd.Flags().SortFlags = false
	statusCmd.Flags().PrintDefaults()
}
