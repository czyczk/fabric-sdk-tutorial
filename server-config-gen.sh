#!/bin/sh

# Generate server config files for org1
for i in {2..15}; do
    # IPFS port: localhost:5001 => localhost:5${i} where ${i} is formatted as "%03g"
    #   e.g.: i = 3 => 5003
    # User name and key file name: User1 => User${i}
    # Port: 8081 => 8${i + 80} where ${i + 80} is formatted as "%03g"
    #   e.g.: i = 3 => 8083
    let port_number=$(expr ${i} + 80)
    printf -v port_number_formatted "%03g" ${port_number}
    printf -v ipfs_port_number_formatted "%03g" ${i}
    sed 's/localhost:5001/localhost:5'${ipfs_port_number_formatted}'/g' ./server.yaml | sed 's/User1/User'${i}'/g' | sed 's/8081/8'${port_number_formatted}'/g' > ./server-u${i}o1.yaml
done

# Generate server config files for org2
for i in {2..15}; do
    # IPFS port: localhost:5017 => localhost:5${i + 16} where ${i + 16} is formatted as "%03g"
    #   e.g.: i = 3 => 5019
    # User name and key file name: User1 => User${i}
    # Port: 8097 => 8${i + 96} where ${i + 96} is formatted as "%03g"
    #   e.g.: i = 3 => 8099
    let port_number=$(expr ${i} + 96)
    printf -v port_number_formatted "%03g" ${port_number}
    let ipfs_port_number=$(expr ${i} + 16)
    printf -v ipfs_port_number_formatted "%03g" ${ipfs_port_number}
    sed 's/localhost:5017/localhost:5'${ipfs_port_number_formatted}'/g' ./server-u1o2.yaml | sed 's/User1/User'${i}'/g' | sed 's/8097/8'${port_number_formatted}'/g' > ./server-u${i}o2.yaml
done