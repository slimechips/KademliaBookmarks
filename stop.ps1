Write-Host "Stopping docker containers..."
docker container stop $(docker container ls -aq)
docker network rm kad_net
docker container prune -f