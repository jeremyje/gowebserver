#!/bin/bash

function BuildBinary() {
	GOOS=$1
	GOARCH=$2
	if [ "$GOOS" == "windows" ]; then
		BINARY_FILE="server.exe"
	else
		BINARY_FILE="server"
	fi
	
	echo "Building ${GOOS} on ${GOARCH}"
	GOOS=${GOOS} GOARCH=${GOARCH} go build server.go
	DEST_DIR="build/${GOOS}-${GOARCH}/"
	mkdir -p ${DEST_DIR}
	mv ${BINARY_FILE} ${DEST_DIR}
}

# Based on table https://golang.org/doc/install/source
BuildBinary "linux" "amd64"
BuildBinary "linux" "386"
BuildBinary "linux" "arm"
BuildBinary "linux" "arm64"
BuildBinary "windows" "amd64"
BuildBinary "windows" "386"
BuildBinary "darwin" "amd64"
BuildBinary "netbsd" "amd64"
BuildBinary "openbsd" "amd64"
BuildBinary "freebsd" "amd64"
BuildBinary "dragonfly" "amd64"

gsutil -m cp -r build/ gs://gowebserver/pub/
gsutil -m -o acl set public-read gs://gowebserver/pub/**
