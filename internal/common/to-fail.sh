#!/bin/bash

echo "Starting"

loop=true

echo "Loop"

while [ "$loop" ]; do
    echo "Starting tests"
    # richgo test -v -race -count=1 -run Broker/broker_closes_all_subs_on_exit ./... || break
    richgo test -v -race -count=1  ./... || break
    # richgo test -v -count=1 -run XXX -benchtime 5s -bench "Reconnect$" ./... || break
done

echo "Done $(date)"
