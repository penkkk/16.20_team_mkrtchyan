#!/usr/bin/env bash
set -euo pipefail

tests_dir="$(cd "$(dirname "$0")" && pwd)"
compose_file="$tests_dir/compose.integration.yml"
project="opd-integration-$$"

cleanup() {
  docker compose -f "$compose_file" -p "$project" down -v --remove-orphans >/dev/null
}
trap cleanup EXIT

docker compose -f "$compose_file" -p "$project" up -d --wait
postgres_port="$(docker compose -f "$compose_file" -p "$project" port postgres 5432 | sed 's/.*://')"
redis_port="$(docker compose -f "$compose_file" -p "$project" port redis 6379 | sed 's/.*://')"

cd "$tests_dir/.."
OPD_TEST_DATABASE_URL="postgres://opd_test:opd_test@127.0.0.1:$postgres_port/opd_test?sslmode=disable" \
OPD_TEST_REDIS_URL="redis://127.0.0.1:$redis_port/0" \
go test -tags=integration -coverpkg=./... ./tests -v
