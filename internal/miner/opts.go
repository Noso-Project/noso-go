package miner

type Opts struct {
	Cpu            int
	IpAddr         string
	IpPort         int
	PoolPw         string
	Wallet         string
	ShowPop        bool
	StatusInterval int
	ExitOnRetry    bool
}
