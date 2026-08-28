GO ?= go
all: test vet build
fmt:
	$(GO) fmt ./...
test:
	GOCACHE=/private/tmp/cubase-agent-gocache $(GO) test ./...
vet:
	GOCACHE=/private/tmp/cubase-agent-gocache $(GO) vet ./...
build:
	GOCACHE=/private/tmp/cubase-agent-gocache $(GO) build ./cmd/cubase-agent
