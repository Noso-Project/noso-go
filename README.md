# Noso Coin Miner

## Quickstart

*NOTE* You should set `-cpu` to the maximum *PHYSICAL* cores on your computer. `go-miner` cannot use hyperthreading/hardware-threads, so setting `-cpu` higher than your *PHYSICAL* cores will likely reduce your overall hashrate.

```
./go-miner-macos \
	-wallet <your wallet address> \
	-cpu <number of CPU cores to use when mining>
```

e.g.
```
./go-miner-macos \
	-wallet Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF \
	-cpu 4
```

## Introduction
`go-miner` is a command line tool for mining the crypto currency [Noso Coin](https://nosocoin.com/). Written using Google's Go language, `go-miner`'s goals are as follows:

* Highly concurrent
* Well optimized
* Cross platform
* Easy to use
