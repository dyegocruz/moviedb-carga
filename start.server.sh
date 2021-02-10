#!/bin/bash

# default is development

if [[ -z "${NODE_ENV}" ]]; then
  echo "Setting to Development"
  export NODE_ENV=development
fi

docker-compose config

docker-compose up -d --build --force