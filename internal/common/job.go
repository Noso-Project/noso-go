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
		for _, w = range HashableSeedChars[:60] {
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

const hashChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var allHashes [][]byte

func AllHashes() [][]byte {
	if allHashes == nil {
		var w, x, y, z rune
		b := make([][]byte, 0)

		for _, w = range hashChars[:10] {
			for _, x = range hashChars {
				for _, y = range hashChars {
					for _, z = range hashChars {
						b = append(b, []byte(fmt.Sprintf("%c%c%c%c", w, x, y, z)))
					}
				}
			}
		}
		allHashes = b
	}

	return allHashes
}

func (j *Job) GenBytes(ctx context.Context) <-chan []byte {
	hashes := AllHashes()
	stream := make(chan []byte, 0)
	go func() {
		defer close(stream)
		for _, h := range hashes {
			select {
			case <-ctx.Done():
				return
			case stream <- h:
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
