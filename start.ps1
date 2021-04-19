$NODES=5
Write-Host "Building new image kadbm"
docker build -t kadbm .
Write-Host "Created Network kad_net"
docker network create --gateway=10.0.0.254 --subnet=10.0.0.0/24 kad_net

for ($i=1; $i -le $NODES; $i++) {
  $port = 9000 + $i
  $ip = "10.0.0." + $i
  Write-Host Creating container $i on $ip, publishing to $port
  docker run -dit `
    --name kad_node_$i `
    --network kad_net `
    -p ${port}:8080 `
    --ip $ip `
    --mount type=bind,source=$pwd/logs/node$i,target=/app/logs `
    kadbm:latest
}

Read-Host -Prompt "Press Enter to Terminate..."

Write-Host "Stopping docker containers..."
docker container stop $(docker container ls -aq)
docker network rm kad_net
docker container prune -f