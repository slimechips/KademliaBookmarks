## Docker-Compose

To launch 5 instances

```bash
docker-compose -f "docker-compose.debug.yml" up -d --scale kademliabookmarks=5
```

## CLI Send UDP Packet

e.g. Connect to UDP port `3000`

```bash
nc -u 127.0.0.1 3000
```