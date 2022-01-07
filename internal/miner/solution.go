package miner

import (
	"context"
	"fmt"
	"sync"

	"github.com/Noso-Project/noso-go/internal/common"
)

func SolutionManager(ctx context.Context, client *common.Client, broker *common.Broker, wg *sync.WaitGroup) {
	wg.Done()

	solStream, err := broker.Subscribe(ctx, common.SolutionTopic)
	if err != nil {
		panic(fmt.Sprint("Got an err and didn't expect one", err))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case sol, ok := <-solStream:
			if !ok {
				return
			}
			// TODO: Need to finalize logger situation
			// common.logger.Debugf("Solution pulled from <-solStream: %v\n", sol)
			// TODO: Log solution found to console
			client.Send(ctx, sol.(common.Solution).String())
		}
	}
}
