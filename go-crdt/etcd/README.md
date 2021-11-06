# etcd cart
An alternative shopping cart implementation using etcd Compare-And-Set (CAS) for atomic additions to the cart.

Set up a 5-node etcd cluster:

```
docker network create etcd-network

REGISTRY=quay.io/coreos/etcd

# For each machine
ETCD_VERSION=v3.5.1
TOKEN=my-etcd-token
CLUSTER_STATE=new
NAME_1=etcd-node-0
NAME_2=etcd-node-1
NAME_3=etcd-node-2
NAME_4=etcd-node-3
NAME_5=etcd-node-4
HOST_1=etcd1
HOST_2=etcd2
HOST_3=etcd3
HOST_4=etcd4
HOST_5=etcd5
CLUSTER=${NAME_1}=http://${HOST_1}:2380,${NAME_2}=http://${HOST_2}:2380,${NAME_3}=http://${HOST_3}:2380,${NAME_4}=http://${HOST_4}:2380,${NAME_5}=http://${HOST_5}:2380
DATA_DIR=/var/lib/etcd

# For node 1
THIS_NAME=${NAME_1}
THIS_IP=${HOST_1}
docker run \
  -d \
  --network etcd-network \
  -p 2379:2379 \
  -p 2380:2380 \
  --name etcd1 ${REGISTRY}:${ETCD_VERSION} \
  /usr/local/bin/etcd \
  --name ${THIS_NAME} \
  --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://0.0.0.0:2380 \
  --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://0.0.0.0:2379 \
  --initial-cluster ${CLUSTER} \
  --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

# For node 2
THIS_NAME=${NAME_2}
THIS_IP=${HOST_2}
docker run \
  -d \
  --network etcd-network \
  --name etcd2 ${REGISTRY}:${ETCD_VERSION} \
  /usr/local/bin/etcd \
  --name ${THIS_NAME} \
  --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://0.0.0.0:2380 \
  --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://0.0.0.0:2379 \
  --initial-cluster ${CLUSTER} \
  --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

# For node 3
THIS_NAME=${NAME_3}
THIS_IP=${HOST_3}
docker run \
  -d \
  --network etcd-network \
  --name etcd3 ${REGISTRY}:${ETCD_VERSION} \
  /usr/local/bin/etcd \
  --name ${THIS_NAME} \
  --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://0.0.0.0:2380 \
  --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://0.0.0.0:2379 \
  --initial-cluster ${CLUSTER} \
  --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

# For node 4
THIS_NAME=${NAME_4}
THIS_IP=${HOST_4}
docker run \
  -d \
  --network etcd-network \
  --name etcd4 ${REGISTRY}:${ETCD_VERSION} \
  /usr/local/bin/etcd \
  --name ${THIS_NAME} \
  --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://0.0.0.0:2380 \
  --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://0.0.0.0:2379 \
  --initial-cluster ${CLUSTER} \
  --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

# For node 5
THIS_NAME=${NAME_5}
THIS_IP=${HOST_5}
docker run \
  -d \
  --network etcd-network \
  --name etcd5 ${REGISTRY}:${ETCD_VERSION} \
  /usr/local/bin/etcd \
  --name ${THIS_NAME} \
  --initial-advertise-peer-urls http://${THIS_IP}:2380 --listen-peer-urls http://0.0.0.0:2380 \
  --advertise-client-urls http://${THIS_IP}:2379 --listen-client-urls http://0.0.0.0:2379 \
  --initial-cluster ${CLUSTER} \
  --initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}
```

```
# check
docker exec -it etcd1 bash
root@28cabc47435a:/# etcdctl member list
2d2ed32011d4b367, started, etcd-node-1, http://etcd2:2380, http://etcd2:2379, false
4269078c6fd53622, started, etcd-node-0, http://etcd1:2380, http://etcd1:2379, false
56e3cc9f3089f54a, started, etcd-node-2, http://etcd3:2380, http://etcd3:2379, false
66c0818f923b94d4, started, etcd-node-4, http://etcd5:2380, http://etcd5:2379, false
7b5e52fc7b83cc73, started, etcd-node-3, http://etcd4:2380, http://etcd4:2379, false
```

```
# fill cart initially
etcdctl get cart
etcdctl put cart []
etcdctl get cart
```

Run cart filler:
```
go build .
# run a container at the same network with mounted dir
docker run --rm --name etcd-runner --network=etcd-network -v /home/nkalyanov/crdt-cart/:/crdt-cart/ -it golang bash
/crdt-cart/go-crdt/etcd/etcd
```

```
# docker stop etcd1 etcd2 etcd3 etcd4 etcd5
# docker rm etcd1 etcd2 etcd3 etcd4 etcd5
```
