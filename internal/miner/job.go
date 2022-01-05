package miner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Noso-Project/noso-go/internal/common"
)

const (
	hashableSeedChars = "!\"#$&')*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^`abcdefghijklmnopqrstuvwxyz{|"
)

func JobManager(ctx context.Context, client *common.Client, wg *sync.WaitGroup) {

	// TODO: Need to wrap this in a for loop and then detect a client reconnect and start over
	//       Might need a sync.Cond from the client for that
	// TODO: Likewise, might need a "ready" WaitGroup in all goroutines to signal readiness?

	builder := newJobBuilder(ctx)

	jobStream := make(chan common.Job, 0)

	jobTopicStream, err := client.Subscribe(common.JobTopic)
	if err != nil {
		panic(err)
	}

	defer client.Unsubscribe(jobTopicStream)

	poolDataStream, err := client.Subscribe(common.PoolDataTopic)
	if err != nil {
		panic(err)
	}
	defer client.Unsubscribe(poolDataStream)

	wg.Done()

	// time.Sleep(time.Second)
	// // Leave jStream nil until we get our first PoolData msg
	var jStream, nJob chan common.Job
	jStream, nJob = nil, nil
	// var once sync.Once
	var job common.Job
loop:
	for {
		select {
		case <-ctx.Done():
			return
		case job = <-nJob:
			nJob = nil
			jStream = jobStream
		case jStream <- job:
			jStream = nil
			nJob = builder.nextJob
		case poolDataMsg := <-poolDataStream:
			// TODO: Only care about POOLSTEPS or JOINOK, can discard PONG
			switch poolDataMsg.(type) {
			case common.Pong:
				continue loop
			}
			// fmt.Printf("Got data from poolDataStream: %v\n", poolDataMsg)
			resp := builder.Update(poolDataMsg)
			if resp == common.JOINOK {
				nJob = builder.nextJob
			}

		case jobTopicMsg := <-jobTopicStream:
			// fmt.Printf("Got jobTopic from jobTopicStream: %v\n", jobTopicMsg.(common.JobStreamReq))
			func(stream <-chan common.Job) {
				select {
				case <-ctx.Done():
					return
				case jobTopicMsg.(common.JobStreamReq).Stream <- stream:
				}
			}(jobStream)
		}
	}
}

func newJobBuilder(ctx context.Context) (j *jobBuilder) {
	j = new(jobBuilder)
	j.ready = make(chan struct{}, 0)
	j.nextJob = make(chan common.Job, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go j.builder(ctx, &wg)
	wg.Wait()

	return
}

type jobBuilder struct {
	ready        chan struct{}
	nextJob      chan common.Job
	poolAddr     string
	seedFromPool string
	targetString string
	targetChars  int
	diff         int
	block        int
	step         int
	poolDepth    int
}

func (j *jobBuilder) builder(ctx context.Context, wg *sync.WaitGroup) {
	var job common.Job

	wg.Done()

	// Wait for a JoinOk
	<-j.ready

	verSha := sha256.Sum256([]byte(common.MinerName))
	verShaHex := hex.EncodeToString(verSha[:])
	ver := verShaHex[:2]

	// TODO: push seed char generation into it's own goroutine
	// Randomize seed chars so that if a miner restarts in the middle of a block,
	// it isn't rehashing already hashed values
	seedChars := []rune(hashableSeedChars)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(seedChars), func(i, j int) { seedChars[i], seedChars[j] = seedChars[j], seedChars[i] })

	for {
		for _, x := range seedChars {
			for _, y := range seedChars {
				for _, z := range seedChars {

					seedBase := j.seedFromPool[:len(j.seedFromPool)-3]
					seed := fmt.Sprintf("%s%c%c%c", seedBase, x, y, z)

					for num := 1; num < 999; num++ {
						postfix := ver + fmt.Sprintf("%03d", num)
						fullSeed := seed + j.poolAddr + postfix
						job = j.newJob(seed, fullSeed)
						select {
						case j.nextJob <- job:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}
}

func (j *jobBuilder) newJob(minerSeedBase, minerSeed string) common.Job {
	// fmt.Println("newJob() Called")
	return common.Job{
		PoolAddr:      j.poolAddr,
		MinerSeedBase: minerSeedBase,
		MinerSeed:     minerSeed,
		TargetString:  j.targetString,
		TargetChars:   j.targetChars,
		Diff:          j.diff,
		Block:         j.block,
		Step:          j.step,
		PoolDepth:     j.poolDepth,
	}
}

// Example job from 1.6.3
// PoolAddr:N6VxgLSpbni8kLbyUAjYXdHCPt2VEp
// SeedMiner:3Q0000###
// SeedPostfix:11001
// SeedFull:3Q0000###N6VxgLSpbni8kLbyUAjYXdHCPt2VEp11001
// SeedFullBytes:[51 81 48 48 48 48 35 35 35 78 54 86 120 103 76 83 112 98 110 105 56 107 76 98 121 85 65 106 89 88 100 72 67 80 116 50 86 69 112 49 49 48 48 49]
// TargetString:b42ff0ec9847e71e77e6b620508f93d0
// TargetChars:11
// Diff:105
// Block:39591
// Step:3
// PoolDepth:3}

func (j *jobBuilder) Update(poolData interface{}) common.ServerMessageType {
	// fmt.Println("Updated() Called")
	switch poolData.(type) {
	case common.JoinOk:
		j.poolAddr = poolData.(common.JoinOk).PoolAddr
		j.seedFromPool = poolData.(common.JoinOk).MinerSeed
		j.targetString = poolData.(common.JoinOk).TargetHash
		j.targetChars = poolData.(common.JoinOk).TargetLen
		j.diff = poolData.(common.JoinOk).Difficulty
		j.block = poolData.(common.JoinOk).Block
		j.step = poolData.(common.JoinOk).CurrentStep
		j.poolDepth = poolData.(common.JoinOk).PoolDepth
		close(j.ready)
		return common.JOINOK
	case common.PoolSteps:
		j.targetString = poolData.(common.PoolSteps).TargetHash
		j.targetChars = poolData.(common.PoolSteps).TargetLen
		j.diff = poolData.(common.PoolSteps).Difficulty
		j.block = poolData.(common.PoolSteps).Block
		j.step = poolData.(common.PoolSteps).CurrentStep
		j.poolDepth = poolData.(common.PoolSteps).PoolDepth
		return common.POOLSTEPS
	}
	return common.OTHER
}

func requestJobStream(ctx context.Context, client *common.Client) <-chan common.Job {

	stream := make(chan (<-chan common.Job), 0)
	client.Publish(common.JobStreamReq{Stream: stream})

	select {
	case s := <-stream:
		return s
	case <-ctx.Done():
		return nil
	}
}
