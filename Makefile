VERSION?="0.0.1"

default: test

test: fmtcheck
	go test ./... -cover -race

testint: fmtcheck
	go test ./... -cover -race -tags=integration -count=1

cover: fmtcheck
	go test ./... -race -coverprofile cover.out -tags=integration && \
	go tool cover -func cover.out && \
	rm cover.out

fmt:
	go fmt ./...

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint: fmtcheck
	golangci-lint run -E golint --exclude-use-default=false --build-tags integration

deeplint: fmtcheck
	golangci-lint run --exclude-use-default=false --enable-all -D dupl --build-tags integration

ci: fmtcheck lint	
	go test ./... -race -cover -tags=integration -coverprofile=coverage.txt -covermode=atomic
	curl -s https://codecov.io/bash | bash -s

local-k8s-portfwd:
	@sh -c "'$(CURDIR)/scripts/k8s_port_forward_local.sh'"

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: default test testint cover fmt fmtcheck lint deeplint ci local-k8s-portfwd