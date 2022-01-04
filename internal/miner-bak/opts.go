package miner

type Opts struct {
	Cpu            int
	IpAddr         string
	IpPort         int
	PoolPw         string
	Wallets        []string
	CurrentWallet  string
	ShowPop        bool
	StatusInterval int
	ExitOnRetry    bool
}
