# Noso Coin Miner

Quick Notes: 

* The noso-go go-miner is currently free, with no fee structure built in. Future versions may include a low percentage fee
* The beta version is hard coded to connect to the noso.dukedog.io pool so I can monitor stats. Future versions will allow the user to select any pool they choose

## Quickstart

Download the latest release from the [Releases Page](https://github.com/leviable/noso-go-releases/releases)

### Quickstart - Windows

1. Unzip the release
2. Edit the `go-miner.bat` file
3. Update the CPU line with the number of PHYSICAL cores you want to use
4. Update the WALLET line with your wallet address
5. Save and close
6. Double click the go-miner.bat file 

### Quickstart - Linux/Mac/ARM

*NOTE* You should set `-cpu` to the maximum *PHYSICAL* cores on your computer. `go-miner` cannot use hyperthreading/hardware-threads, so setting `-cpu` higher than your *PHYSICAL* cores will likely reduce your overall hashrate.

```
./go-miner-macos \
	-wallet <your wallet address> \
	-cpu <number of CPU cores to use when mining>
```

e.g.
```
./go-miner-macos -wallet Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF -cpu 4
```

## Introduction
`go-miner` is a command line tool for mining the crypto currency [Noso Coin](https://nosocoin.com/). Written using Google's Go language, `go-miner`'s goals are as follows:

* Highly concurrent
* Well optimized
* Cross platform
* Easy to use

## Understanding the output

Future version of `go-miner` will have a more user friendly output. For now, you should only need to pay attention to the PONG lines:

```
<- PONG PoolData 5351 5AFADEC0006675E408E5C06AA09C0120 10 6 99 953841173 -5 336517
```

* Block: 5351
* Current Step: 6
* Difficulty: 99
* Balance: 9.53841173 Noso
* Blocks Till Payment: 5
* Pool HashRate: 336.517 MegaHashes/second