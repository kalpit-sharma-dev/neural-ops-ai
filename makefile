.PHONY: tidy build test test-integration test-coverage seed seed-fast loadtest frontend-e2e

tidy:
	cd backend && go mod tidy

build:
	cd backend && go build ./...

build-collector:
	cd backend && go build -o bin/collector ./cmd/collector

test:
	cd backend && go test ./...

test-coverage:
	bash scripts/coverage-gate.sh

test-integration:
	cd backend && go test -tags=integration ./tests/integration/... -count=1 -timeout=20m

seed:
	cd backend && go run ./scripts/seed

seed-fast:
	cd backend && go run ./scripts/seed -logs 5000 -skip-http

loadtest:
	cd backend && go run ./scripts/loadtest -rate 100000 -duration 60s -workers 64 -batch 200

loadtest-k6:
	k6 run tests/load/k6/observability.js

contract-test:
	cd backend && go test ./tests/contract/... -count=1

openapi-test:
	cd backend && go test ./tests/contract/... -run 'TestOpenAPI|TestGatewayOpenAPI' -count=1

package-deb:
	bash deploy/packaging/build-deb.sh

package-rpm:
	bash deploy/packaging/build-rpm.sh

gen-mtls-certs:
	bash scripts/gen-mtls-certs.sh

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-e2e:
	cd frontend && npm run test:e2e

compose-build:
	bash scripts/docker-build-stack.sh

compose-up:
	bash scripts/quickstart.sh

compose-down:
	docker compose -f infra/docker-compose.yml down

migrate-up:
	migrate -path backend/migrations -database "$${POSTGRES_DSN:-postgres://neuralops:neuralops@localhost:5432/neuralops?sslmode=disable}" up

migrate-down:
	migrate -path backend/migrations -database "$${POSTGRES_DSN:-postgres://neuralops:neuralops@localhost:5432/neuralops?sslmode=disable}" down 1
