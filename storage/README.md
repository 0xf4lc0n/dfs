# Get started

## Run database:

```bash
docker run --rm -p 5433:5432 -e "POSTGRES_PASSWORD=postgres" -e "POSTGRES_DB=dfs_storage" --name pg_storage postgres:latest
```

## Run RabbitMq

```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.10-management
```
