#!/bin/bash

# default is development

if [[ -z "${GO_ENV}" ]]; then
  echo "Setting to Development"
  export GO_ENV=development
fi

docker-compose config

docker-compose up -d --build --force --remove-orphans

docker-compose -f docker-compose-catalog-worker.yml up -d --build --scale catalog-process-worker=4
