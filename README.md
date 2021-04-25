# Kademlia Bookmarks

## Running binary

```bash
go build main
# windows
./main.exe <my_assigned_node_id> <ip_address_of_bootstrap> <republish duration> <expiry_duration>
``` 

## View in webpage

Node 1 will be on `localhost:9001`, Node 2 will be on `localhost:9002` etc

## Docker

### (New) Build image and start docker containers

```powershell
./start.ps1
```

Containers will delete themselves if you press enter to terminate.

### (New) Start docker containers without building

```powershell
./fastStart.ps1
```

Containers will delete themselves if you press enter to terminate.

### (New) Stop docker containers

In case you didnt stop the previous two scripts properly

```powershell
./stop.ps1
```

### Docker-Compose

First off I recommend deleting all existing containers. (Probably dont need run this all the time)

```
docker-compose rm -f
```

Make sure you are in the project folder. Modify the number of replicas you want (default I set to 4) in the `replicas:` option in the `docker-compose.debug.yml`. Then run

```bash
docker-compose -f "docker-compose.debug.yml" up -d --build
```

Or you can just use the `Docker: Compose Up` option from VSCode Command Palette and choose the `docker-compose.debug.yml` if you have docker extension installed

To stop them

```bash
docker-compose -f "docker-compose.debug.yml" down
```

Or you can just right click on the containers tab in VSCode docker extension and `Compose Down`

Note that the container names ARE NOT the same as the node id of the server running in the container. It is all based on IP Address.


### View Logs for each Docker Container

Vscode Docker Extension: Right click on container and `Attach Shell`

CLI: `docker exec -it <FULL_CONTAINER_ID> sh`

Then in the container's shell, `tail -f -n 1000 app.log`

### Docker Admin (Just for testing purposes)

### CLI Send UDP Packet (IP Address Method)

Network nodes are defined on `172.16.238.0/24`.

E.g. to connect to node on `172.16.238.2`

```bash
nc -u 172.16.238.2 1053
```

#### List Docker Networks

```bash
docker network ls
```

#### Inspect Docker Network

See node names and ip addresses on network. (Network name is `kademliabookmarks_kademlia_net`)

```bash
docker network inspect kad_net
```

#### Jepsen Tests
https://github.com/zackteo/kademlia.bookmarks
