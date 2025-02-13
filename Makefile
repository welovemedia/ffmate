version := $(shell cat .version)

prepare:
	go mod tidy

test:
	go test ./...

dev: 
	go run -race main.go server -d "*" --send-telemetry=false

mkdir+bin:
	mkdir -p _bin

build+frontend:
	cd ui && pnpm i && pnpm run generate

build: build+frontend mkdir+bin 
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o _bin/darwin-arm64 main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o _bin/darwin-amd64 main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-musl-gcc   go build -ldflags "-s -w -linkmode external -extldflags "-static"" -o _bin/linux-arm64 main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc go build -ldflags "-s -w -linkmode external -extldflags "-static"" -o _bin/linux-amd64 main.go

changelog:
	auto-changelog --output CHANGELOG.md

swagger:
	swag init --outputTypes go

update: build
	rm -rf _update
	go-selfupdate -o=_update/ffmate _bin/ $(version)
	aws s3 sync _update s3://ffmate/_update --profile cloudflare-r2 --delete

release: update
	git tag -a v$(version) -m "v$(version)"
	GITHUB_TOKEN=$$(cat ~/.config/goreleaser/github_token_ffmate) goreleaser release --clean

air: 
	air
