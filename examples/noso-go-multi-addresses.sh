#!/usr/bin/env bash

# #########################################
# 
# Please update POOL, WALLET, and CPU below
# 
# #########################################

# Valid Pools:
#   devnoso
#   dukedog.io
#   mining.moe
#   russiapool
#   yzpool

# TODO: Check ENVs for pool/wallet/cpu
#
# Example values:
# POOL=devnoso
# WALLET=N3nCJEtfWSB77HHv2tFdKGL7onevyDg
# CPU=4

POOL=devnoso
CPU=1
WALLET="leviablefarmy1"
WALLET+=" leviablefarmy2"
WALLET+=" leviablefarmy3"
WALLET+=" leviablefarmy4"
WALLET+=" leviablefarmy5"
WALLET+=" leviablefarmy6"
WALLET+=" leviablefarmy7"
WALLET+=" leviablefarmy8"
WALLET+=" leviablefarmy9"
WALLET+=" leviablefarmy10"
WALLET+=" leviablefarmy11"
WALLET+=" leviablefarmy12"
WALLET+=" leviablefarmy13"
WALLET+=" leviablefarmy14"

# #########################################
# 
# No user editable code below
# 
# #########################################

while true; do
  for wallet in ${WALLET:?Variable not set or is empty}; do
    ./bin/noso-go-linux-amd64 mine pool \
        "${POOL:?Variable not set or is empty}" \
        --wallet "$wallet" \
        --cpu ${CPU:?Variable not set or is empty} \
        --exit-on-retry
  done
done
