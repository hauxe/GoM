#!/bin/bash
cd "$(dirname "$0")"
docker-compose -f docker-zipkin/docker-compose.yml -f docker-zipkin/docker-compose-cassandra.yml -f docker-zipkin/docker-compose-ui.yml -f docker-zipkin/docker-compose-kafka10.yml up -d