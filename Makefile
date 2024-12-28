VERSION?="0.0.1"

.PHONY: default
default: test

.PHONY: test
test: fmtcheck
	go test ./... -cover -race

.PHONY: testint
testint: fmtcheck
	go test ./... -cover -race -tags=integration -count=1

.PHONY: cover
cover: fmtcheck
	go test ./... -coverpkg=./... -coverprofile=cover.out -tags=integration -covermode=atomic && \
	go tool cover -func=cover.out &&\
	rm cover.out

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: fmtcheck
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: lint
lint: fmtcheck
	docker run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint -v run

.PHONY: deeplint
deeplint: fmtcheck
	docker run --env=GOFLAGS=-mod=vendor --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint run --exclude-use-default=false --enable-all -D dupl --build-tags integration

.PHONY: deps-start
deps-start:
	docker compose up -d

.PHONY: deps-stop
deps-stop:
	docker compose down

.PHONY: ci
ci: fmtcheck
	go test ./... -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic
	
# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL: