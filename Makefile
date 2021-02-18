SHELL = /bin/bash -o pipefail
export COMPOSE_DOCKER_CLI_BUILD = 1
export DOCKER_BUILDKIT = 1

build:
	@docker-compose build

up:
	@docker-compose up --detach --renew-anon-volumes --remove-orphans

down:
	@docker-compose down --remove-orphans

logs:
	@docker-compose logs -f

help:
	@docker run --rm backblaze-server:1.0.0 --help

go-build: CGO_ENABLED=0
go-build: GOARCH=amd64
go-build: GOOS=linux
go-build:
	@go build -mod=vendor -o bin/backblaze-server-amd64 ./app/backblaze-server/...

sinclude .env # silent include; no error if file is not yet decrypted (same as -include) 
run:
	@./bin/backblaze-server-amd64 \
		--domain files.127.0.0.1.nip.io \
		--backblaze-application-key ${B2SERVER_BACKBLAZE_APPLICATION_KEY} \
		--backblaze-key-id ${B2SERVER_BACKBLAZE_KEY_ID}

sops-encrypt:
	$$(go env GOPATH)/bin/sops --pgp E03741868C02382277A3D0A141611E17EA20D77B -e .env > .enc.env

sops-decrypt:
	$$(go env GOPATH)/bin/sops -d .enc.env > .env
