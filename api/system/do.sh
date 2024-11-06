#!/bin/bash
APP_NAME=${APP_NAME:-system}

function build_docker () {
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o docker/$APP_NAME-api ./main.go ./plugins.go

	docker build --tag=rxc/$APP_NAME:k8s --no-cache -f ./docker/Dockerfile .
}

# function start(){
#     go run main.go
# } 

# start

build_docker