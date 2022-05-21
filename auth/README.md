# Get started

## Run database:

```bash
docker run --rm -p 5432:5432 -e "POSTGRES_PASSWORD=postgres" -e "POSTGRES_DB=dfs_auth" --name pg_auth postgres:latest
```

## Run RabbitMq

```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.10-management
```

## Run auth service

Create .env file in the `auth` directory with the following content:

```
DB_CONNECTION_STRING="host=localhost user=postgres password=postgres dbname=postgres port=5432"
```

Build dfs-auth image and run a container:

```bash
# Inside auth folder
docker image build --tag dfs-auth .
docker run -p 8080:8080 dfs-auth
```

Or run service without docker:

```bash
go run .\main.go
```
