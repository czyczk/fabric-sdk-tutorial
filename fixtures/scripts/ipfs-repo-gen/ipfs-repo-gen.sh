#!/bin/bash
readonly REPRESENTITIVE_PEER_ADDR="\/ip4\/ipfs0\/tcp\/4001\/ipfs\/QmeUNYGNzYVfGsBZ6H6yXtp1LSvrNDYVdYc7Du1jNCpBJT"

for i in {4..31}; do
    export IPFS_PATH=./ipfs-template/ipfs$i
    ipfs init
    cp ./ipfs-template/ipfs0/swarm.key $IPFS_PATH/swarm.key
    ipfs bootstrap rm --all
    sed -i 's/\"Bootstrap\": null/\"Bootstrap\": \[\n    \"'$REPRESENTITIVE_PEER_ADDR'\"\n  \]/g' $IPFS_PATH/config
done
