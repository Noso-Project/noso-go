package miner

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

type Opts struct {
	Cpu    int
	IpAddr string
	IpPort int
	PoolPw string
	Wallet string
}

func GetOpts() *Opts {
	ipAddr := flag.String("addr", "", "IP address of Noso Pool")
	ipPort := flag.Int("port", 8082, "IP port of the Noso Pool. Defaults to 8082")
	poolPw := flag.String("password", "", "Password for the NosoPool")
	wallet := flag.String("wallet", "", "Noso address from your wallet")
	cpu := flag.Int("cpu", 0, "Number of CPUs to use when mining. Defaults to MAXCPU - 1")

	flag.Parse()

	if *ipAddr == "" {
		fmt.Println("-addr cannot be blank")
		os.Exit(1)
	} else if *poolPw == "" {
		fmt.Println("-password cannot be blank")
		os.Exit(1)
	} else if *wallet == "" {
		fmt.Println("-wallet cannot be blank")
		os.Exit(1)
	}

	if *cpu == 0 {
		c := runtime.NumCPU() - 1
		cpu = &c
	}

	return &Opts{
		IpAddr: *ipAddr,
		IpPort: *ipPort,
		PoolPw: *poolPw,
		Wallet: *wallet,
		Cpu:    *cpu,
	}
}
