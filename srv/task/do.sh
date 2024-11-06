#!/bin/bash
APP_NAME=${APP_NAME:-task}

function build_docker () {
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o docker/task-srv ./main.go ./plugins.go

	docker build --tag=rxc/$APP_NAME:k8s --no-cache -f ./docker/Dockerfile .
}

build_docker