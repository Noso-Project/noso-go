# Noso Coin Miner

[![Build and Tests Status](https://github.com/Noso-Project/noso-go/workflows/noso-go/badge.svg?branch=main)](https://github.com/Noso-Project/noso-go/actions)
[![Supports Windows](https://img.shields.io/badge/support-Windows-blue?logo=Windows)](https://github.com/Noso-Project/noso-go/releases/latest)
[![Supprts Linux](https://img.shields.io/badge/support-Linux-yellow?logo=Linux)](https://github.com/Noso-Project/noso-go/releases/latest)
[![Supports macOS](https://img.shields.io/badge/support-macOS-black?logo=macOS)](https://github.com/Noso-Project/noso-go/releases/latest)
[![License](https://img.shields.io/github/license/Noso-Project/noso-go)](https://github.com/Noso-Project/noso-go/blob/master/LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/Noso-Project/noso-go?label=latest%20release)](https://github.com/Noso-Project/noso-go/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/Noso-Project/noso-go/total)](https://github.com/Noso-Project/noso-go/releases)

## Quickstart

Download the latest release for your OS and architecture from the [Releases Page](https://github.com/Noso-Project/noso-go/releases)

In most instances you will want the `64` version unless you have a fairly old computer

### Quickstart - Windows

1. Unzip the release
2. Edit the `noso-go.bat` file
3. Update the CPU line with the number of PHYSICAL cores you want to use
4. Update the WALLET line with your wallet address
5. Save and close
6. Double click the noso-go.bat file

NOTE: To find out how many physical cores you have, you can run the following command in the command prompt (works on Windows 2000/NT or later):
```
wmic computersystem get numberoflogicalprocessors
```

### Quickstart - Linux/Mac/ARM
Untar the archive:

```
$ tar -zxvf noso-go-v0.8.0-linux-amd64.tgz
x noso-go
```

*NOTE* You should set `--cpu` to the maximum *PHYSICAL* cores on your computer. `go-miner` cannot use hyperthreading/hardware-threads, so setting `--cpu` higher than your *PHYSICAL* cores will likely reduce your overall hashrate.

```
./noso-go mine \
    --address <IP Address or URL of the Noso pool> \
    --port <Port of the Noso pool>
    --password <Password for the Noso pool> \
    --wallet <Your wallet address> \
    --cpu <Number of CPU cores to use when mining>
```

e.g.

```
./noso-go mine \
    --address noso.dukedog.io \
    --port 8082 \
    --password duke \
    --wallet Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF \
    --cpu 4
```

NOTE: If you get a permissions error, run this command in your terminal window, then try again:
```
chmod a+x noso-go
```

### Quickstart - Android

1. Download the latest 'linux-arm64' (or 'linux-arm') tar archive, and extract its contents to a local folder on your Android device.
2. From the Google Play Store, Install [RS File Manager](https://play.google.com/store/apps/details?id=com.rs.explorer.filemanager)
3. From the Google Play Store, Install [Termux](https://play.google.com/store/apps/details?id=com.termux)
4. Open RS File Manager and browse to the location where you placed your noso-go arm file(s)
5. Press and hold on the noso-go file, select the menu and choose "Open As", then select "Termux" (Optional: In Termux, you can type ```ls``` to list your arm file(s). This verifies that Termux started in the right folder.
6. To start the noso-go miner, use the following command:
```
./noso-go mine --address <pool IP or URL> --port <Pool's Port Number> --password <Pool's Password> --wallet <Your Noso Wallet Address> --cpu <Number of physical cores on your CPU>
```

Note: When using all physical cores on your Android device, you may want to make sure it is in a well-cooled area so that heat does not become a problem.

## Introduction
`noso-go` is a command line tool for mining the crypto currency [Noso Coin](https://nosocoin.com/). Written using Google's Go language, `noso-go`'s goals are as follows:

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
