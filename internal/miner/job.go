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

	// TODO: This is here because the JobManager needs to be listening to JobTopic
	//       stream before JoinOk is received. Should probably do this another way
	wg.Done()

	var job, nilJob common.Job
	jStream := jobStream
	nJob := builder.nextJob
	nJob, jStream = nil, nil
loop:
	for {
		if job == nilJob {
			nJob = builder.nextJob
			jStream = nil
		} else {
			nJob = nil
			jStream = jobStream
		}
		select {
		case <-ctx.Done():
			fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			return
		case jStream <- job:
			job = nilJob
		case job = <-nJob:
		case poolDataMsg := <-poolDataStream:
			fmt.Println("DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")
			switch poolDataMsg.(type) {
			case common.Pong:
				continue loop
			}
			job = nilJob
			builder.Update(poolDataMsg)

		case jobTopicMsg := <-jobTopicStream:
			fmt.Println("EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE")
			func(stream <-chan common.Job) {
				select {
				case <-ctx.Done():
					return
				// TODO: Deadlock/Livelock possible, probably need to timeout here
				case jobTopicMsg.(common.JobStreamReq).Stream <- stream:
				case <-time.After(100 * time.Millisecond):
				}
			}(jobStream)
		}
	}
}

func newJobBuilder(ctx context.Context) (j *jobBuilder) {
	j = new(jobBuilder)
	j.joined = make(chan struct{}, 0)
	j.newBlock = make(chan struct{}, 0)
	j.nextJob = make(chan common.Job, 0)
	j.mu = new(sync.Mutex)

	var wg sync.WaitGroup
	wg.Add(1)
	go j.builder(ctx, &wg)
	wg.Wait()

	return
}

type jobBuilder struct {
	joined       chan struct{}
	newBlock     chan struct{}
	nextJob      chan common.Job
	poolAddr     string
	seedFromPool string
	targetString string
	targetChars  int
	diff         int
	block        int
	step         int
	poolDepth    int
	mu           *sync.Mutex
}

func (j *jobBuilder) builder(ctx context.Context, wg *sync.WaitGroup) {
	var job common.Job

	wg.Done()

	// Wait for the manager to receive and process a JoinOk
	<-j.joined

	// Rough approach for including the noso-go version within the hash
	// string that is ultimately written to the blockchain
	// If not an official build, this should be "11"
	verSha := sha256.Sum256([]byte(common.MinerName))
	verShaHex := hex.EncodeToString(verSha[:])
	ver := verShaHex[:2]

	// TODO: push seed char generation into it's own goroutine

	seedPostfixStream := make(chan string, 0)
	go seedCharGen(ctx, seedPostfixStream)

	jobCtx, jobCancel := context.WithCancel(ctx)

seedLoop:
	for {
		seedBase := j.seedFromPool[:len(j.seedFromPool)-3]
		seed := fmt.Sprintf("%s%s", seedBase, <-seedPostfixStream)

		for num := 1; num < 999; num++ {
			postfix := ver + fmt.Sprintf("%03d", num)
			fullSeed := seed + j.poolAddr + postfix
			job = j.newJob(jobCtx, seed, fullSeed)
			select {
			case <-j.newBlock:
				// Cancel old jobs
				fmt.Println("111111111111111111111111111111111111111111111111111111")
				jobCancel()
				fmt.Println("111111111111111111111111111111111111111111111111111111")
				jobCtx, jobCancel = context.WithCancel(ctx)
				fmt.Println("111111111111111111111111111111111111111111111111111111")
				continue seedLoop
			case j.nextJob <- job:
				fmt.Println("3333333333333333333333333333333333333")
			case <-ctx.Done():
				return
			}
		}
	}
}

func (j *jobBuilder) newJob(ctx context.Context, minerSeedBase, minerSeed string) (job common.Job) {
	j.mu.Lock()
	defer j.mu.Unlock()
	job = common.NewJob(ctx)
	job.PoolAddr = j.poolAddr
	job.MinerSeedBase = minerSeedBase
	job.MinerSeed = minerSeed
	job.TargetString = j.targetString
	job.TargetChars = j.targetChars
	job.Diff = j.diff
	job.Block = j.block
	job.Step = j.step
	job.PoolDepth = j.poolDepth

	return
}

// Example job from 1.6.2
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
	switch poolData.(type) {
	case common.JoinOk:
		func() {
			j.mu.Lock()
			defer j.mu.Unlock()
			j.poolAddr = poolData.(common.JoinOk).PoolAddr
			j.seedFromPool = poolData.(common.JoinOk).MinerSeed
			j.targetString = poolData.(common.JoinOk).TargetHash
			j.targetChars = poolData.(common.JoinOk).TargetLen
			j.diff = poolData.(common.JoinOk).Difficulty
			j.block = poolData.(common.JoinOk).Block
			j.step = poolData.(common.JoinOk).CurrentStep
			j.poolDepth = poolData.(common.JoinOk).PoolDepth
		}()
		close(j.joined)
		return common.JOINOK
	case common.PoolSteps:
		var oldBlock int
		func() {
			j.mu.Lock()
			defer j.mu.Unlock()
			oldBlock = j.block
			j.targetString = poolData.(common.PoolSteps).TargetHash
			j.targetChars = poolData.(common.PoolSteps).TargetLen
			j.diff = poolData.(common.PoolSteps).Difficulty
			j.block = poolData.(common.PoolSteps).Block
			j.step = poolData.(common.PoolSteps).CurrentStep
			j.poolDepth = poolData.(common.PoolSteps).PoolDepth
		}()

		if oldBlock != j.block {
			j.newBlock <- struct{}{}
		}
		return common.POOLSTEPS
	}
	return common.OTHER
}

func seedCharGen(ctx context.Context, stream chan string) {

	// Randomize seed chars so that if a miner restarts in the middle of a block,
	// it isn't rehashing already hashed values
	seedChars := []rune(hashableSeedChars)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(seedChars), func(i, j int) { seedChars[i], seedChars[j] = seedChars[j], seedChars[i] })

	for {
		for _, x := range seedChars {
			for _, y := range seedChars {
				for _, z := range seedChars {
					select {
					case <-ctx.Done():
						return
					case stream <- fmt.Sprintf("%c%c%c", x, y, z):
					}
				}
			}
		}
	}
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
