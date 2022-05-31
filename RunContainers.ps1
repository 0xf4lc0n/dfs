Write-Host "Running postgres database for Auth microservice"
docker run -d --rm -p 5432:5432 -e "POSTGRES_PASSWORD=postgres" -e "POSTGRES_DB=dfs_auth" --name pg_auth postgres:latest

Write-Host "Running postgres database for Storage microservice"
docker run -d --rm -p 5433:5432 -e "POSTGRES_PASSWORD=postgres" -e "POSTGRES_DB=dfs_storage" --name pg_storage postgres:latest

Write-Host "Running postgres database for Share microservice"
docker run -d --rm -p 5434:5432 -e "POSTGRES_PASSWORD=postgres" -e "POSTGRES_DB=dfs_share" --name pg_share postgres:latest

Write-Host "Running postgres database for Sharespace microservice"
docker run -d --rm -p 5435:5432 -e "POSTGRES_PASSWORD=postgres" -e "POSTGRES_DB=dfs_sharespace" --name pg_sharespace postgres:latest

Write-Host "Running RabbitMq"
docker run -d --rm -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:3.10-management
