# Noso Coin Miner

[![Build and Tests Status](https://github.com/Noso-Project/noso-go/workflows/noso-go/badge.svg?branch=main)](https://github.com/Noso-Project/noso-go/actions)
[![Supports Windows](https://img.shields.io/badge/support-Windows-blue?logo=Windows)](https://github.com/Noso-Project/noso-go/releases/latest)
[![Supprts Linux](https://img.shields.io/badge/support-Linux-yellow?logo=Linux)](https://github.com/Noso-Project/noso-go/releases/latest)
[![Supports macOS](https://img.shields.io/badge/support-macOS-black?logo=macOS)](https://github.com/Noso-Project/noso-go/releases/latest)
[![License](https://img.shields.io/github/license/Noso-Project/noso-go)](https://github.com/Noso-Project/noso-go/blob/master/LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/Noso-Project/noso-go?label=latest%20release)](https://github.com/Noso-Project/noso-go/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/Noso-Project/noso-go/total)](https://github.com/Noso-Project/noso-go/releases)

## Quickstart

* [Windows](docs/quickstart-windows.md)
* [Linux](docs/quickstart-linux.md)
* [MacOs](docs/quickstart-macos.md)
* [Raspberry Pi](docs/quickstart-raspberrypi.md)
* [Android](docs/quickstart-android.md)

## Introduction
`noso-go` is a command line tool for mining the cryptocurrency [Noso Coin](https://nosocoin.com/). Written using Google's Go language, `noso-go`'s goals are as follows:

* Free to use
* Highly concurrent
* Well optimized
* Cross platform
* Easy to use

`noso-go` is currently confirmed to run on the following platforms

* Windows (32 and 64 bit)
* Linux (32 and 64 bit)
* MacOS (64 bit)
* Raspberry Pi (arm64)
* Google Pixel 2 (arm64)
* Google Pixel 5 (arm64)

## Understanding the output

Future version of `noso-go` will have a more user friendly output. For now, you should only need to pay attention to the PING and PONG lines:

```
-> PING 4954
```

* Your Miner's Hash Rate: 4,954 KiloHashes/second, or ~5 MH/s

```
<- PONG PoolData 5351 5AFADEC0006675E408E5C06AA09C0120 10 6 99 953841173 -5 336517
```

* Block: 5351
* Current Step: 6
* Difficulty: 99
* Balance: 9.53841173 Noso
* Blocks Till Payment: 5
* Pool HashRate: 336.517 MegaHashes/second

A `status` message will display every 60 seconds:

```
************************************
Miner Status
Miner's Wallet Addr : leviable3
Current Block       : 33383
Miner Hash Rate     :  55.302 Mhash/s
Pool Hash Rate      :   1.632 Ghash/s
Pool Balance        : 22.40475776 Noso
Blocks Till Payment : -20

Proof of Participation
----------------------
PoP Sent            : 813
PoP Accepted        : 808
************************************
```

## Benchmarking

Coming soon

## Chrome/Windows/MacOS Warnings

When downloading the release, you will probably get a warning from your browser, operating system and/or anti-virus software that the release is unsafe. This is because, as of this writing, this project is unable to sign the binaries with trusted certificates, so your browser/OS/AV immediately detects it as an unsigned binary and flags it as a potential threat. You have a couple options to overcome this:

1. First and foremost: inspect the code yourself and make sure you are comfortable with it
2. Build the binary yourself, and your OS wont complain about it. See the [Building](#Building) section below for more info
3. Instruct your browser/OS/AV that you accept the risks
   - Chrome:
     - Click the ^ next to `Discard` and select `Keep`
       ![](images/chrome-keep.png)
   - Windows MSE
     - (Not recommended) Turn off real-time protection:
       ![](images/mse-real-time.png)
     - (Recommended) Create an exclusion zone for noso-go releases, and download them to that location:
       ![](images/mse-excluded-locations.png)
   - MacOS
     - The first time your run the binary you will get a popup like so. Click `Cancel`:
       ![](images/mac-1-popup.png)
     - Open your `System Preferences` app and click on the `Security & Privacy` icon
     - There should be a warning on the bottom about the `noso-go` application being blocked. Click the `Allow Anyway` button:
       ![](images/mac-2-allow-anyways.png)
   - Linux
     - So far I have seen no reports of any flavor of Linux complaining about the binaries. If you come across a problem, please open an [Issue](https://github.com/Noso-Project/noso-go/issues) in this repo and I will add it to the README

## Building

### Prerequisites

* The [Go Compiler](https://golang.org/dl/) (I am using go1.16.3, however most older versions should work fine)

### Steps

1. Download the source code or clone this repo
2. Determine your target OS and Architecture
   - OS options are: `windows`, `linux` or `darwin`
   - Architecture options are: `386`, `amd64`, `arm`, or `arm64`
3. Compile (various examples below):
   - Windows: ```$ GOOS=windows GOARCH=amd64 go build -o noso-go main.go```
   - MacOS: ```$ GOOS=darwin GOARCH=amd64 go build -o noso-go main.go```
   - Linux: ```$ GOOS=linux GOARCH=amd64 go build -o noso-go main.go```
   - ARM: ```$ GOOS=linux GOARCH=arm64 go build -o noso-go main.go```
