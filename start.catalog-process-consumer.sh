#!/bin/bash

# default is development

if [[ -z "${GO_ENV}" ]]; then
  echo "Setting to Development"
  export GO_ENV=development
fi

docker-compose config

docker-compose -f docker-compose-catalog-process.yml up --build -d
