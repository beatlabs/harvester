VERSION?="0.0.1"
DOCKER = docker

default: test

test: fmtcheck
	go test ./... -cover -race

testint: fmtcheck deps
	go test ./... -cover -race -tags=integration -count=1
	docker stop harvester-consul
	docker stop harvester-redis


cover: fmtcheck
	go test ./... -coverpkg=./... -coverprofile=cover.out -tags=integration -covermode=atomic && \
	go tool cover -func=cover.out &&\
	rm cover.out

fmt:
	go fmt ./...

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint: fmtcheck
	$(DOCKER) run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint -v run

deeplint: fmtcheck
	$(DOCKER) run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run --exclude-use-default=false --enable-all -D dupl --build-tags integration

deps:
	docker container inspect harvester-consul > /dev/null 2>&1 || docker run -d --rm -p 8500:8500 -p 8600:8600/udp --name=harvester-consul consul:1.4.3 agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0  -http-port 8500 -log-level=err
	docker container inspect harvester-redis > /dev/null 2>&1 || docker run -d --rm  -p 6379:6379 --name harvester-redis -e ALLOW_EMPTY_PASSWORD=yes bitnami/redis:6.2

deps-stop:
	docker stop harvester-consul
	docker stop harvester-redis

ci: fmtcheck lint deps
	go test ./... -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic
	docker stop harvester-consul
	docker stop harvester-redis

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: default test testint cover fmt fmtcheck lint deeplint ci deps deps-stop