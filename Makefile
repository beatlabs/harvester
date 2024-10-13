VERSION?="0.0.1"

default: test

test: fmtcheck
	go test ./... -cover -race

testint: fmtcheck
	go test ./... -cover -race -tags=integration -count=1
	
cover: fmtcheck
	go test ./... -coverpkg=./... -coverprofile=cover.out -tags=integration -covermode=atomic && \
	go tool cover -func=cover.out &&\
	rm cover.out

fmt:
	go fmt ./...

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint: fmtcheck
	docker run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint -v run

deeplint: fmtcheck
	docker run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint run --exclude-use-default=false --enable-all -D dupl --build-tags integration

deps-start:
	docker compose up -d

deps-stop:
	docker compose down

ci: fmtcheck
	go test ./... -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic
	
# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: default test testint cover fmt fmtcheck lint deeplint ci deps-start deps-stop