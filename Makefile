# Local development. Requires Go 1.25+ on your PATH.
# Building the web UI additionally requires Node.js + npm.
.PHONY: build build-go run tidy web release-check release-snapshot

# Build the embedded web UI into internal/server/dist. The output is committed
# so `go build`/`go install` work without a Node toolchain; run this whenever
# the web/ sources change.
web:
	npm --prefix web install
	npm --prefix web run build

build: web
	go build -o tagbackup ./cmd/tagbackup

# Build the Go binary using the already-built (committed) web assets.
build-go:
	go build -o tagbackup ./cmd/tagbackup

run:
	go run ./cmd/tagbackup

tidy:
	go mod tidy

release-check:
	goreleaser check

release-snapshot:
	goreleaser release --snapshot --clean
