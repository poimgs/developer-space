#!/usr/bin/env bash
set -euo pipefail

# Seed an admin user in the production database.
# Usage: bash scripts/seed-admin.sh --email admin@example.com --name "Admin Name"

cd "$(dirname "$0")/.."
docker compose -f docker-compose.prod.yml exec api /api seed-admin "$@"
