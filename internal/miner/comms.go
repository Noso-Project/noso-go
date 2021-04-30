package miner

import "time"

func NewComms() *Comms {
	return &Comms{
		PoolAddr:          make(chan string, 0),
		MinerSeed:         make(chan string, 0),
		TargetString:      make(chan string, 0),
		TargetChars:       make(chan int, 0),
		Block:             make(chan int, 0),
		Step:              make(chan int, 0),
		Diff:              make(chan int, 0),
		Balance:           make(chan string, 0),
		BlocksTillPayment: make(chan int, 0),
		StepSolved:        make(chan int, 0),
		HashRate:          make(chan int, 0),
		Jobs:              make(chan Job, 0),
		Reports:           make(chan Report, 0),
		Solutions:         make(chan Solution, 0),
		Joined:            make(chan struct{}, 0),
		ClearHashRate:     make(chan struct{}, 0),
	}
}

type Comms struct {
	PoolAddr          chan string
	MinerSeed         chan string
	TargetString      chan string
	TargetChars       chan int
	Block             chan int
	Step              chan int
	Diff              chan int
	Balance           chan string
	BlocksTillPayment chan int
	StepSolved        chan int
	HashRate          chan int
	Jobs              chan Job
	Reports           chan Report
	Solutions         chan Solution
	Joined            chan struct{}
	ClearHashRate     chan struct{}
}

type Report struct {
	WorkerNum string
	Hashes    int
	Duration  time.Duration
}
