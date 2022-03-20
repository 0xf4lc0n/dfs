# Get started

Create .env file in the `auth` directory with the following content:

```
DB_CONNECTION_STRING="host=dfs_db_postgres user=postgres password=postgres dbname=postgres port=5432"
```

Create .env file in the `dfs` directory with the following content:

```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
```

Run docker compose:

```bash
docker-compose up -d --build
```
