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

go-build: CGO_ENABLED=0
go-build: GOARCH=amd64
go-build: GOOS=linux
go-build:
	@go build -mod=vendor -o bin/backblaze-server-amd64 ./app/backblaze-server/...
