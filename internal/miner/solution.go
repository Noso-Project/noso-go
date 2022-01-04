package miner

import (
	"context"
	"fmt"
	"sync"

	"github.com/Noso-Project/noso-go/internal/common"
)

func SolutionManager(ctx context.Context, client *common.Client, wg *sync.WaitGroup) {
	wg.Done()

	solStream, err := client.Subscribe(common.SolutionTopic)
	if err != nil {
		panic(fmt.Sprint("Got an err and didn't expect one", err))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case sol := <-solStream:
			// TODO: Need to finalize logger situation
			// common.logger.Debugf("Solution pulled from <-solStream: %v\n", sol)
			client.Send(sol.(common.Solution).String())
		}
	}
}
