#!/usr/bin/env bash
set -euo pipefail

docker compose -f docker/docker-compose.yml --env-file .env up -d --build
