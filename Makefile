# Local development. Requires Go 1.25+ on your PATH.
.PHONY: build run tidy release-check release-snapshot

build:
	go build -o tagbackup ./cmd/tagbackup

run:
	go run ./cmd/tagbackup

tidy:
	go mod tidy

release-check:
	goreleaser check

release-snapshot:
	goreleaser release --snapshot --clean
