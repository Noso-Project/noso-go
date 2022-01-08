package common

import (
	"context"
	"fmt"
)

func NewJob(ctx context.Context) Job {
	return Job{
		done: ctx.Done(),
	}
}

type Job struct {
	done          <-chan struct{}
	PoolAddr      string
	MinerSeedBase string
	MinerPostfix  string
	MinerSeed     string
	TargetString  string
	TargetChars   int
	Diff          int
	Block         int
	Step          int
	PoolDepth     int
}

func (j *Job) Gen(ctx context.Context) <-chan string {
	var w, x, y, z rune
	stream := make(chan string, 0)
	go func() {
		defer close(stream)
		// 5 was chosen so that it would take roughly 1 second to iterate
		// through all the hashes on one modern-ish cpu thread
		for _, w = range HashableSeedChars[:5] {
			for _, x = range HashableSeedChars {
				for _, y = range HashableSeedChars {
					for _, z = range HashableSeedChars {
						stream <- fmt.Sprintf("%c%c%c%c", w, x, y, z)
					}
				}
			}
		}
	}()
	return stream
}

func (j *Job) Done() <-chan struct{} {
	return j.done
}

type JobStreamReq struct {
	Stream chan (<-chan Job)
}
