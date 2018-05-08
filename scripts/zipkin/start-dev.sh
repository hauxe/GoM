#!/bin/bash
cd "$(dirname "$0")"
docker-compose -f docker-zipkin/docker-compose.yml -f docker-zipkin/docker-compose-ui.yml up -d