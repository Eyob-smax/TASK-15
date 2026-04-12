#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COVERAGE=0

for arg in "$@"; do
  case "$arg" in
    --coverage) COVERAGE=1 ;;
    *) echo "Unknown flag: $arg" >&2; exit 1 ;;
  esac
done

echo "========================================"
echo "FitCommerce Test Suite Runner"
echo "========================================"
echo ""

cd "$SCRIPT_DIR"

# ----------------------------------------
# Backend: unit tests + API tests
# ----------------------------------------
echo "--- Backend Tests ---"
if [ "$COVERAGE" -eq 1 ]; then
  BACKEND_CMD="go mod tidy && go mod download && \
    go test ./unit_tests/... -v -count=1 -coverprofile=coverage_unit.out -covermode=atomic && \
    go tool cover -func=coverage_unit.out | tail -1 && \
    go test ./api_tests/... -v -count=1 -timeout=30m -coverprofile=coverage_api.out -covermode=atomic && \
    go tool cover -func=coverage_api.out | tail -1"
else
  BACKEND_CMD="go mod tidy && go mod download && \
    go test ./unit_tests/... -v -count=1 && \
    go test ./api_tests/... -v -count=1 -timeout=30m"
fi
docker compose --profile test run --rm backend-test sh -c "$BACKEND_CMD"
echo ""

# ----------------------------------------
# Frontend: unit tests
# ----------------------------------------
echo "--- Frontend Tests ---"
if [ "$COVERAGE" -eq 1 ]; then
  FRONTEND_CMD="(npm ci --no-audit --no-fund || npm install --no-audit --no-fund) && \
    npx vitest run --config vitest.config.ts --coverage"
else
  FRONTEND_CMD="(npm ci --no-audit --no-fund || npm install --no-audit --no-fund) && \
    npx vitest run --config vitest.config.ts"
fi
docker compose --profile test run --rm frontend-test sh -c "$FRONTEND_CMD"
echo ""

echo "========================================"
echo "ALL TEST SUITES PASSED"
echo "========================================"
