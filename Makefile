GITREF=main
IMGTAG=dev

build:
	go build ./...

run-indexer: build
	go run cmd/indexer/main.go --env=./docker/dev/docker.env

generate-contracts:
	@./scripts/contractGen.sh

db-migrate:
	tern migrate -c db/tern.conf -m db

install:
	go install ./cmd/...

docker-build:
	docker build --no-cache --pull --platform=linux/amd64 --build-arg=GITREF=${GITREF} --rm -f Dockerfile -t airdroputils:${IMGTAG} .

docker-tag:
	docker tag airdroputils:dev ghcr.io/arkeonetwork/airdroputils:${IMGTAG}

docker-push:
	docker push ghcr.io/arkeonetwork/airdroputils:${IMGTAG}

push-image: docker-build docker-tag docker-push
