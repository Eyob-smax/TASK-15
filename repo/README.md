# FitCommerce Quick Start

Minimal instructions to run the project locally with Docker.

## Prerequisites

- Docker Desktop (with `docker compose`)

## Startup

From this `repo/` directory:

```bash
docker compose up --build
```

## Access

- Frontend HTTPS: https://localhost:3443
- Frontend HTTP (redirects to HTTPS): http://localhost:3000
- Backend API (TLS): https://localhost:8080
- PostgreSQL: localhost:5432

## Run Tests

From this `repo/` directory:

```bash
./run_tests.sh
```

With coverage:

```bash
./run_tests.sh --coverage
```

## Stop

```bash
docker compose down
```

## Clean Reset (optional)

Removes containers, networks, and volumes (deletes local DB data):

```bash
docker compose down -v
```
