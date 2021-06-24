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
	"net"
	"os"
	"sort"
	"strings"

	"github.com/Noso-Project/noso-go/internal/miner"
	"github.com/spf13/cobra"
)

var (
	list     bool
	info     bool
	pools    map[string]*miner.Opts
	poolOpts = &miner.Opts{}
)

// poolCmd represents the pool command
var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Connect to a named Noso pool and mine Noso coin",
	Long: `Connect to a named Noso pool and mine Noso coin
Example usage:

List available pools
./noso-go mine pool --list

List info about a specific pool
./noso-go mine pool yzpool --info

Start mining with a pool
./noso-go mine pool devnoso    --wallet <your wallet address>
./noso-go mine pool dukedog.io --wallet <your wallet address>
./noso-go mine pool mining.moe --wallet <your wallet address>
./noso-go mine pool russiapool --wallet <your wallet address>
./noso-go mine pool yzpool     --wallet <your wallet address>
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if list {
			return nil
		}
		if len(args) < 1 {
			return errors.New("requires a pool name (e.g. 'noso-go mine pool yzpool')")
		}
		poolName := strings.ToLower(args[0])
		if _, ok := pools[poolName]; !ok {
			return errors.New("Unrecognized pool name. Use 'noso-go mine pool --list' for list of pools")
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

		if poolOpts.Cpu < 1 {
			cmd.PrintErrln("Error: --cpu cannot be less than 1")
			os.Exit(1)
		}

		ipAddr, err := net.LookupIP(pool.IpAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get IP address for domain: %v\n", err)
			os.Exit(1)
		}
		poolOpts.IpAddr = ipAddr[0].String()
		poolOpts.IpPort = pool.IpPort
		poolOpts.PoolPw = pool.PoolPw

		miner.Mine(poolOpts)
	},
}

func init() {
	loadPools()
	mineCmd.AddCommand(poolCmd)

	poolCmd.Flags().BoolVarP(&list, "list", "l", false, "List known pool names")
	poolCmd.Flags().BoolVarP(&info, "info", "i", false, "Print Pool information and exit")
	poolCmd.Flags().StringVarP(&poolOpts.Wallet, "wallet", "w", "", "Noso wallet address to send payments to")
	poolCmd.Flags().IntVarP(&poolOpts.Cpu, "cpu", "c", 4, "Number of CPU cores to use")
	poolCmd.Flags().BoolVarP(&poolOpts.ShowPop, "show-pop", "", false, "Show PoP solutions in output")

	poolCmd.Flags().SortFlags = false
	poolCmd.Flags().PrintDefaults()
}

func printPoolInfo(poolName string, poolOpts *miner.Opts) {
	msg := `Pool info for %s:
	Pool Address : %s
	Pool Port    : %d
	Pool Password: %s
`
	fmt.Printf(msg, poolName, poolOpts.IpAddr, poolOpts.IpPort, poolOpts.PoolPw)
}

func listPools() {
	poolNames := []string{}
	for pool, _ := range pools {
		poolNames = append(poolNames, pool)
	}
	sort.Strings(poolNames)

	nameList := strings.Join(poolNames, "\n\t- ")
	fmt.Printf("Please use one of the following pool names:\n\t- %s\n", nameList)
}

func loadPools() {
	// TODO: support loading a pools config file at runtime too
	// TODO: Use github.com/markbates/pkger to package a Yaml
	//       file instead of hard coding these here
	pools = make(map[string]*miner.Opts)
	pools["dukedog.io"] = &miner.Opts{
		IpAddr: "noso.dukedog.io",
		IpPort: 8082,
		PoolPw: "duke",
	}
	pools["dukedogio"] = &miner.Opts{
		IpAddr: "noso.dukedog.io",
		IpPort: 8082,
		PoolPw: "duke",
	}
	pools["dukedog"] = &miner.Opts{
		IpAddr: "noso.dukedog.io",
		IpPort: 8082,
		PoolPw: "duke",
	}
	pools["russiapool"] = &miner.Opts{
		IpAddr: "95.54.44.147",
		IpPort: 8082,
		PoolPw: "RussiaPool",
	}
	pools["mining.moe"] = &miner.Opts{
		IpAddr: "Node1.mining.moe",
		IpPort: 8082,
		PoolPw: "miningmoe",
	}
	pools["miningmoe"] = &miner.Opts{
		IpAddr: "Node1.mining.moe",
		IpPort: 8082,
		PoolPw: "miningmoe",
	}
	pools["devnoso"] = &miner.Opts{
		IpAddr: "DevNosoEU.nosocoin.com",
		IpPort: 8082,
		PoolPw: "UnMaTcHeD",
	}
	pools["yzpool"] = &miner.Opts{
		IpAddr: "noso.yzpool.gold",
		IpPort: 8082,
		PoolPw: "YZpool",
	}
}
