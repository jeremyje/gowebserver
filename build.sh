#!/bin/bash
# Builds binaries for each OS/Architecture.

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

function BuildDefaultBinaries() {
	BuildBinary "linux" "amd64"
	BuildBinary "linux" "arm"
	BuildBinary "linux" "386"
	BuildBinary "windows" "amd64"
	BuildBinary "windows" "386"
}

# Based on table https://golang.org/doc/install/source

function BuildExtendedBinaries() {
	BuildBinary "linux" "arm64"
	BuildBinary "darwin" "amd64"
	BuildBinary "netbsd" "amd64"
	BuildBinary "openbsd" "amd64"
	BuildBinary "freebsd" "amd64"
	BuildBinary "dragonfly" "amd64"
}

function PrepareGithubRelease() {
	mv build/linux-amd64/server build/server-amd64
	mv build/linux-386/server build/server-386
	mv build/linux-arm/server build/server-arm
	mv build/windows-amd64/server.exe build/windows-amd64.exe
	mv build/windows-386/server.exe build/windows-386.exe
}

BuildDefaultBinaries
if [ "$1" == "all" ]; then
	BuildExtendedBinaries
fi

if [ "$1" == "release" ]; then
	PrepareGithubRelease
fi
