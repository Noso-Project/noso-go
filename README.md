# Noso Coin Miner

## Quickstart

```
./go-miner-macos \
	-addr <pool ip address> \
	-port <pool ip port> \
	-password <pool password> \
	-wallet <your wallet address> \
	-cpu <number of CPU cores to use when mining>
```

e.g.
```
./go-miner-macos \
	-addr noso.dukedog.io \
	-port 8082 \
	-password duke \
	-wallet Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF \
	-cpu 4
```

## Introduction
`go-miner` is a command line tool for mining the crypto currency [Noso Coin](https://nosocoin.com/). Written using Google's Go language, `go-miner`'s goals are as follows:

* Highly concurrent
* Well optimized
* Cross platform
* Easy to use