#!/bin/bash
GO_DIR=$(pwd)
echo ${GO_DIR}
GitClone() {
    BASE_DIR=vendor/$1
    REPO=$2
    echo ${GO_DIR}
    cd ${GO_DIR}
    rm -rf ${BASE_DIR}
    echo git submodule add ${REPO} ${BASE_DIR} --force
    git submodule add ${REPO} ${BASE_DIR} --force
    cd ${GO_DIR}
}

GithubClone() {
    USER_PATH=$1
    BASE_DIR=github.com/${USER_PATH}
    REPO=git@github.com:${USER_PATH}.git
    GitClone ${BASE_DIR} ${REPO}
}

BitbucketClone() {
    USER_PATH=$1
    BASE_DIR=bitbucket.org/${USER_PATH}
    REPO=git@bitbucket.org:${USER_PATH}.git
    GitClone ${BASE_DIR} ${REPO}
}

GithubClone beorn7/perks
GithubClone golang/protobuf
GithubClone matttproud/golang_protobuf_extensions
GithubClone prometheus/client_golang
GithubClone prometheus/client_model
GithubClone prometheus/common
GithubClone prometheus/procfs
GithubClone rs/cors
GithubClone rs/xhandler
GithubClone stretchr/testify
GitClone golang.org/x/net https://go.googlesource.com/net
GithubClone go-yaml/yaml
