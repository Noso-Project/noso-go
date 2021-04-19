#!/bin/bash

# Pool: DevNoso
#ADDRESS=23.95.233.179
#PORT=8084
#PASSWORD=UnMaTcHeD

# Pool: nosopoolDE
#ADDRESS=199.247.3.186
#PORT=8082
#PASSWORD=nosopoolDE

# Pool: YZpool
#ADDRESS=81.68.115.175
#PORT=8082
#PASSWORD=YZpool

# Pool: Hodl
#ADDRESS=104.168.99.254
#PORT=8082
#PASSWORD=Hodl

# Pool: DogFaceDuke
ADDRESS=noso.dukedog.io
PORT=8082
PASSWORD=duke

WALLET=[Your Wallet Address]
CPU=[Number of CPUs]

while true; do
  ./noso-go mine \
    --address "${ADDRESS}" \
    --port "${PORT}" \
    --password "${PASSWORD}" \
    --wallet "${WALLET}" \
    --cpu "${CPU}"
done
