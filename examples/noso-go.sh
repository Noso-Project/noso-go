#!/bin/bash

# #########################################
# 
# Please update POOL, WALLET, and CPU below
# 
# #########################################

# Valid Pools:
#   devnoso
#   dukedock.io
#   hodl
#   nosopoolde
#   russiapool
#   yzpool

# TODO: Check ENVs for pool/wallet/cpu
#
# Example values:
# POOL=yzpool
# WALLET=N3nCJEtfWSB77HHv2tFdKGL7onevyDg
# CPU=4

POOL=
WALLET=
CPU=

# #########################################
# 
# No user editable code below
# 
# #########################################

while true; do
  ./noso-go mine pool \
    "${POOL:?Variable not set or is empty}" \
    --wallet "${WALLET:?Variable not set or is empty}" \
    --cpu ${CPU:?Variable not set or is empty}

  if [ "$?" != "0" ]; then
      exit 1
  fi
done
