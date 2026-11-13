#!/bin/bash

set -e

export COMPOSE_HTTP_TIMEOUT="${COMPOSE_HTTP_TIMEOUT:-300}"

# default is development

if [[ -z "${GO_ENV}" ]]; then
  echo "Setting to Development"
  export GO_ENV=development
fi

docker-compose config

docker-compose up -d --build --force-recreate --remove-orphans

docker-compose -f docker-compose-catalog-worker.yml up -d --build --scale catalog-process-worker=8
