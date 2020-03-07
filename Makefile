DOCKER_ENV_NAME := app

init: docker/build docker/up

up: docker/up

down: docker/down

deps: docker/deps

run: docker/run

sh: docker/sh

local/deps:
	which dep || go get -v -u github.com/golang/dep/cmd/dep
	dep ensure

local/run:
	go run main.go

docker/build: docker-compose.yml
	@docker-compose build

docker/up:
	@docker-compose up -d

docker/down:
	@docker-compose down

docker/deps:
	@docker-compose exec $(DOCKER_ENV_NAME) make local/deps

docker/run:
	@docker-compose exec $(DOCKER_ENV_NAME) make local/run

docker/sh:
	@docker-compose exec $(DOCKER_ENV_NAME) /bin/bash

.PHONY: init up down deps run sh
.PHONY: local/* docker/*
