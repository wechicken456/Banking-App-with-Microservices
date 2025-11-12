#!/bin/sh
# to be used inside another docker container whose job is to set up a proxy
echo "Setting up 50ms DB latency via Toxiproxy..."

curl -X POST http://toxiproxy:8474/populate -s -H "Content-Type: application/json" -d '[{
  "name": "postgres",
  "listen": "0.0.0.0:5432",
  "upstream": "account-db:5432",
  "enabled": true
}]'

curl -X POST http://toxiproxy:8474/proxies/postgres/toxics -H "Content-Type: application/json" -d '{
  "name": "latency",
  "type": "latency",
  "stream": "downstream",
  "toxicity": 1.0,
  "attributes": { "latency": 50, "jitter": 5 }
}'
