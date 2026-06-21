.DEFAULT_GOAL := test

.PHONY: test testint cover fmt fmtcheck lint deeplint deps-start deps-stop ci

test: fmtcheck
	go test ./... -cover -race

testint: fmtcheck
	go test ./... -cover -race -tags=integration -count=1

cover: fmtcheck
	go test ./... -coverpkg=./... -coverprofile=cover.out -tags=integration -covermode=atomic
	go tool cover -func=cover.out
	rm cover.out

fmt:
	go fmt ./...

fmtcheck:
	./scripts/gofmtcheck.sh

lint: fmtcheck
	docker run --env=GOFLAGS=-mod=vendor --rm -v "$(CURDIR)":/app -w /app golangci/golangci-lint:v2.11.1 golangci-lint -v run

deeplint: fmtcheck
	docker run --env=GOFLAGS=-mod=vendor --rm -v "$(CURDIR)":/app -w /app golangci/golangci-lint:v2.11.1 golangci-lint run --exclude-use-default=false --enable-all -D dupl --build-tags integration

deps-start:
	docker compose up -d

deps-stop:
	docker compose down

ci: fmtcheck
	go test ./... -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic
