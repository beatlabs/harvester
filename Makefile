VERSION?="0.0.1"

default: test

test: fmtcheck
	go test ./... -cover -race

testint: fmtcheck
	go test ./... -cover -race -tags=integration -count=1

test-with-coverage:
	go test ./... -mod=vendor -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic
	./scripts/codecov.sh

cover: fmtcheck
	go test ./... -coverpkg=./... -coverprofile=cover.out -tags=integration -covermode=atomic && \
	go tool cover -func=cover.out &&\
	rm cover.out

fmt:
	go fmt ./...

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint: fmtcheck
	golangci-lint run -E golint,gofmt,unparam,goconst,prealloc,stylecheck,unconvert --exclude-use-default=false --build-tags integration

deeplint: fmtcheck
	golangci-lint run --exclude-use-default=false --enable-all -D dupl --build-tags integration

ci: fmtcheck lint	
	go test ./... -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic

ci-initialize:
	docker-compose -f ./docker-compose.ci.yaml build

ci-lint:
	docker-compose -f ./docker-compose.ci.yaml run harvester-ci make lint

ci-test:
	docker-compose -f ./docker-compose.ci.yaml run -e CODECOV_TOKEN=${CODECOV_TOKEN} harvester-ci make test-with-coverage

ci-all:
	make ci-initialize
	make ci-lint
	make ci-test

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: default test testint cover fmt fmtcheck lint deeplint ci
