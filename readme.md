# Kademlia Bookmarks

## Docker

### Docker-Compose

Make sure you are in the project folder. Modify the number of replicas you want (default I set to 4) in the `replicas:` option in the `docker-compose.debug.yml`. Then run

```bash
docker-compose -f "docker-compose.debug.yml" up -d --build
```

To stop them

```bash
docker-compose -f "docker-compose.debug.yml" down
```

### Connect to bash terminal in a container

Get the `<container-id>` from inspect docker network

```bash
docker exec -it <container-id> bash
```

### CLI Send UDP Packet

#### DNS Method

Containers are named `kademliaboomarks_node_<id>`. Starts from 1.

e.g. Connect to UDP port `1053` for node 1.

```bash
nc -u kademliabookmarks_node_1 1053
```

#### IP Address Method

Network nodes are defined on `172.16.238.0/24`.

E.g. to connect to node on `172.16.238.2`

```bash
nc -u 172.16.238.2 1053
```

### Docker Admin

#### List Docker Networks

```bash
docker network ls
```

#### Inspect Docker Network

See node names and ip addresses on network. (Network name is `kademliabookmarks_kademlia_net`)

```bash
docker network inspect kademliabookmarks_kademlia_net
```