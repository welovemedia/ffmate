VERSION := $(shell cat .version)
DOCKER_TAG=latest
DOCKER_REPO=ghcr.io/welovemedia/ffmate
.PHONY: e2e

prepare:
	go mod tidy

test:
	go test -race ./...

test+coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...

test+coverage+func:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...

lint:
	golangci-lint run --timeout=5m ./...

dev+sqlite:
	go run -race main.go server --identifier="sev.moovit.de" --tray=true --debug="info:?,debug:?,warn:?,error:?" --send-telemetry=false --no-ui=true --database="${FFMATE_DB_SQLITE}" --labels="dev" --port 2999

dev+sqlite+memory:
	go run -race main.go server --identifier="sev.moovit.de" --tray=false --debug="info:?,debug:?,warn:?,error:?" --send-telemetry=false --no-ui=true --database=":memory:"

dev+postgres:
	go run -race main.go server --identifier="sev.moovit.de" --tray=false --debug="*" --send-telemetry=false --no-ui=true --database="${FFMATE_DB_POSTGRES}" --labels="dev"

dev+postgres+2:
	go run -race main.go server --identifier="sev-2.moovit.de" --tray=false --debug="*" --send-telemetry=false --no-ui=true --database="${FFMATE_DB_POSTGRES}" --port 2999 --labels="dev2"

dev+e2e:
	rm -rf /tmp/test.db
	go run -race main.go server --identifier="test.e2e" --tray=false --debug="info:?,debug:?,warn:?,error:?" --send-telemetry=false --no-ui=true --database="/tmp/test.db"

swagger:
	swag init --outputTypes go -o internal/docs

mkdir+bin:
	mkdir -p _bin

build+frontend:
	mv internal/controller/ui/ui-build/index.html index.html
	cd ui && pnpm i && pnpm run generate
	cp -r ui/.output/public/* internal/controller/ui/ui-build/

build: test swagger build+frontend mkdir+bin build+darwin build+linux build+windows

build+darwin+only: test swagger build+frontend mkdir+bin build+darwin
	rm -rf internal/controller/ui/ui-build/*
	mv index.html internal/controller/ui/ui-build/

build+linux+only: test swagger build+frontend mkdir+bin build+linux
	rm -rf internal/controller/ui/ui-build/*
	mv index.html internal/controller/ui/ui-build/

build+darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o _bin/darwin-arm64 main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o _bin/darwin-amd64 main.go

build+linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-musl-gcc go build -ldflags "-s -w -linkmode external -extldflags "-static"" -o _bin/linux-arm64 main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc go build -ldflags "-s -w -linkmode external -extldflags "-static"" -o _bin/linux-amd64 main.go

build+windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o _bin/windows-arm64.exe main.go
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o _bin/windows-amd64.exe main.go

build+app: build
	cp _bin/darwin-arm64 _app/ffmate_arm64/ffmate.app/Contents/MacOS/ffmate
	cp _bin/darwin-amd64 _app/ffmate_amd64/ffmate.app/Contents/MacOS/ffmate

docker+build:
	docker buildx build -f Dockerfile.amd64 -t ${DOCKER_REPO}:${VERSION}-amd64 --platform linux/amd64 --load .
	docker buildx build -f Dockerfile.arm64 -t ${DOCKER_REPO}:${VERSION}-arm64 --platform linux/arm64 --load .

docker+push:
	docker push ${DOCKER_REPO}:${VERSION}-amd64
	docker push ${DOCKER_REPO}:${VERSION}-arm64

docker+manifest:
	docker manifest create ${DOCKER_REPO}:${VERSION} --amend ${DOCKER_REPO}:${VERSION}-amd64 ${DOCKER_REPO}:${VERSION}-arm64
	docker manifest push ${DOCKER_REPO}:${VERSION}
	docker manifest create ${DOCKER_REPO}:latest --amend ${DOCKER_REPO}:${VERSION}-amd64 ${DOCKER_REPO}:${VERSION}-arm64
	docker manifest push ${DOCKER_REPO}:latest

docker+release: docker+build docker+push docker+manifest

update: build
	rm -rf _update
	go-selfupdate -o=_update/ffmate _bin/ $(VERSION)
	aws s3 sync _update s3://ffmate/_update --profile cloudflare-r2 --delete --checksum-algorithm=CRC32

air:
	air -c .air.toml
