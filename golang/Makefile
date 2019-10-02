IMAGE?="ot101:golang"

default: build

run:
	go run ./...

build:
	go build -o ot101 ./...

.PHONY: docker.build
docker.build:
	docker build -t $(IMAGE) .

.PHONY: docker.run
docker.run: docker.build
	docker run -e OT_AZ=$(OT_AZ) -p 8085:8085 -p 8087:8087 -it $(IMAGE)

.PHONY: lint
lint:
	golangci-lint run
