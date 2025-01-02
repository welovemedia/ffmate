version := $(shell cat .version)

prepare:
	go mod tidy

test:
	go test ./...

dev: 
	go run -race main.go

mkdir+bin:
	mkdir -p _bin

build: mkdir+bin
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o _bin/darwin-arm64 main.go

update: build
	rm -rf _update
	go-selfupdate -o=_update/ffmate _bin/ $(version)
	aws s3 sync _update s3://ffmate/_update --profile cloudflare-r2 --delete

release:
	git tag -a v$(version) -m "v$(version)"
	goreleaser release --clean

air: 
	air
