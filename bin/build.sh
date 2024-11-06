#!/bin/bash

dir=`pwd`

build() {
	for d in $(ls ./../$1); do
		echo "building $1/$d"
		cd $dir && cd ../ && cd $1/$d
		# echo "$dir/$1/$d"
		CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w'
	done
}

# build 
# build srv
build web
