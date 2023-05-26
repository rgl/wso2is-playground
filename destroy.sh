#!/bin/bash
set -euo pipefail

# destroy the existing environment.
#docker compose run --workdir /host destroy
docker compose --profile test down --volumes
docker compose --profile client down --volumes
docker compose down --volumes
rm -f terraform.{log,tfstate,tfstate.backup} tfplan
