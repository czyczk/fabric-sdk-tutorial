#!/bin/sh

# Normalize filenames of secret keys from ***_sk to sk
for org_dir in ./crypto-config/peerOrganizations/*/ ; do
    # Normalize filenames of secret keys for peers
    for peer_dir in ${org_dir}peers/*/ ; do
        keystore_path=${peer_dir}"msp/keystore/"
        for sk_path in ${keystore_path}*_sk ; do
            mv ${sk_path} ${keystore_path}sk
        done
    done

    # Normalize filenames of secret keys for users
    for peer_dir in ${org_dir}users/*/ ; do
        keystore_path=${peer_dir}"msp/keystore/"
        for sk_path in ${keystore_path}*_sk ; do
            mv ${sk_path} ${keystore_path}sk
        done
    done
done