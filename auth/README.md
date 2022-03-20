# Get started

## Run database:

```bash
docker run --rm -p 5432:5432 -e "POSTGRES_PASSWORD=postgres" --name pg postgres:latest
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
