#!/usr/bin/env bash

# #########################################
# 
# Please update POOL, WALLET, and CPU below
# 
# #########################################

# Valid Pools:
#   devnoso
#   leviable
#   russiapool

# Example values:
# POOL="devnoso"
# WALLET="devteam_donations"
# CPU=4

# You can specify multiple wallet addresses, which will be
# cycled through round-robin after each disconnect. Useful
# If you want to maintain a single shell script for multiple
# miners (Note you MUST suround wallet addresses with quotes:
# WALLET="leviable leviabl2 leviable3"

POOL="devnoso"
WALLET=""
CPU=1

# #########################################
# 
# No user editable code below
# 
# #########################################

echo $WALLET

for wallet in ${WALLET:?Variable not set or is empty}; do
    wallets+=" --wallet $wallet"
done

while true; do
  ./noso-go mine pool \
    "${POOL:?Variable not set or is empty}" \
    $wallets \
    --cpu ${CPU:?Variable not set or is empty}

  if [ "$?" != "0" ]; then
      exit 1
  fi
done
