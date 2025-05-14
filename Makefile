VERSION := $(shell cat .version)
DOCKER_TAG=latest
DOCKER_REPO=ghcr.io/welovemedia/ffmate
.PHONY: e2e

prepare:
	go mod tidy

test:
	go test ./...
test+slow:
	go test ./... --tags=slow

e2e:
	go run -race e2e/main.go server --send-telemetry=false --database="file::memory:?cache=shared" --loglevel=none

dev: 
	go run -race main.go server -d "*" --send-telemetry=false

mkdir+bin:
	mkdir -p _bin

build+frontend:
	cd ui && pnpm i && pnpm run generate

build: test swagger build+frontend mkdir+bin 
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o _bin/darwin-arm64 main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o _bin/darwin-amd64 main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-musl-gcc go build -ldflags "-s -w -linkmode external -extldflags "-static"" -o _bin/linux-arm64 main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc go build -ldflags "-s -w -linkmode external -extldflags "-static"" -o _bin/linux-amd64 main.go

build+app: build
	cp _bin/darwin-arm64 _app/ffmate_arm64/ffmate.app/Contents/MacOS/ffmate
	cp _bin/darwin-amd64 _app/ffmate_amd64/ffmate.app/Contents/MacOS/ffmate

docker+build:
	docker buildx build -f Dockerfile.amd64 -t ${DOCKER_REPO}:${VERSION}-amd64 -t ${DOCKER_REPO}:latest --platform linux/amd64 --load .
	docker buildx build -f Dockerfile.arm64 -t ${DOCKER_REPO}:${VERSION}-arm64 --platform linux/arm64 --load .

docker+push: 
	docker push ${DOCKER_REPO}:${VERSION}-amd64
	docker push ${DOCKER_REPO}:${VERSION}-arm64
	docker push ${DOCKER_REPO}:latest

docker+manifest:
	docker manifest create ${DOCKER_REPO}:${VERSION} --amend ${DOCKER_REPO}:${VERSION}-amd64 ${DOCKER_REPO}:${VERSION}-arm64
	docker manifest push ${DOCKER_REPO}:${VERSION}

docker+release: docker+build docker+push docker+manifest

changelog:
	auto-changelog --output CHANGELOG.md

swagger:
	swag init --outputTypes go

update: build
	rm -rf _update
	go-selfupdate -o=_update/ffmate _bin/ $(VERSION)
	aws s3 sync _update s3://ffmate/_update --profile cloudflare-r2 --delete

release: update
	git tag -a v$(VERSION) -m "v$(VERSION)"
	GITHUB_TOKEN=$$(cat ~/.config/goreleaser/github_token_ffmate) goreleaser release --clean
	$(MAKE) docker+release

air: 
	air
