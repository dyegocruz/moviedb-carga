# Moviedb Charge

[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/golang.org/x/example)

This repository contains a Go program to get data from the [TMDB API](https://developer.themoviedb.org/docs "TMDB API") and insert/update in a MongoDb and an Elasticsearch instance to improve the search.

The main tasks (after the first charge) is to maintain a property database updated with the same data on the [TMDB API](https://developer.themoviedb.org/docs "TMDB API"). Using the daily export files to check if the catalog is complete and inserting new registers not found on the Charge database.

OBS.: The first charge needs to be used out of the schedule because it takes a long long time.

## Server startup

Run `GO_ENV=production ./start.server.sh` to build and start the catalog and eight
worker replicas. Worker scaling is controlled by `--scale` in the script, not by
Swarm's `deploy.replicas`.

The script stops on the first failure and defaults `COMPOSE_HTTP_TIMEOUT` to 300
seconds for Docker Compose v1. Override it for slower servers, for example:

```sh
GO_ENV=production COMPOSE_HTTP_TIMEOUT=600 ./start.server.sh
```

If requests still time out, inspect `docker info`, `docker stats --no-stream`,
`docker-compose ps`, and
`docker-compose -f docker-compose-catalog-worker.yml ps` on the server. Check
Docker daemon health and available CPU, memory, and disk space before increasing
the timeout further. Compose v2 (`docker compose`) does not use this variable.
