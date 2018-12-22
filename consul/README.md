# consul setup

## intro
`consul` is a distributed key-value store and a service-mesh solution that can be used for observability of
services running on different raspberry pi's (clients) in the cluster. The cluster setup described here
consists of one `server` node and three `client` nodes.

Services running on each client only listen on `localhost` or `127.0.0.1`, so they are not reachable
across the wire. Consul `sidecar` approach is then used to link these services across the network. The
service and the sidecar are able to talk to each other on the localhost, but only sidecars talk to each
other over the network and such communication is encrypted via mutual TLS (`mTLS`).

## pre-requisite
Following components need to be installed on each of the participating nodes:
* consul binary
* all the services that will run on the node

Transfer the contents of `server/xy` folder to the `xy` server node. Similarly, transfer contents of
`client/xy` folder to the `xy` client node.

## start server
`ssh` onto the server node and run the bash scripts from the folder that you just copied over. See if
`consul ui` is up and running by visiting http://server-ip:8500

## start clients
`ssh` onto each of the client nodes and run bash scripts from the folder that you just copied over. Visit
the consul up again and see that all of the services are up and running.