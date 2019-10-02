OT_AZ?=1
IMAGE?="ot101:java"

default: build

.PHONY: run
run:
	OT_AZ=$(OT_AZ) mvn spring-boot:run

.PHONY: build
build:
	mvn -B dependency:resolve

.PHONY: docker.build
docker.build:
	docker build -t $(IMAGE) .

.PHONY: docker.run
docker.run: docker.build
	docker run -e OT_AZ=$(OT_AZ) -p 8085:8085 -p 8087:8087 -it $(IMAGE)
