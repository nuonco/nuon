.PHONY: build client test

client:
	cd client && bun install --frozen-lockfile && bun run build

build: client
	go build -o bin/bundle-portal .

test: client
	go test ./... -count=1
