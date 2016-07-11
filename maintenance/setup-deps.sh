#!/bin/bash
GO_DIR=$(pwd)
echo ${GO_DIR}
GitClone() {
    BASE_DIR=vendor/$1
    REPO=$2
    echo ${GO_DIR}
    cd ${GO_DIR}
    git submodule add ${REPO} ${BASE_DIR}
    cd ${GO_DIR}
}

GithubClone() {
    USER_PATH=$1
    BASE_DIR=github.com/${USER_PATH}
    REPO=https://github.com/${USER_PATH}.git
    GitClone ${BASE_DIR} ${REPO}
}

BitbucketClone() {
    USER_PATH=$1
    BASE_DIR=bitbucket.org/${USER_PATH}
    REPO=https://bitbucket.org/${USER_PATH}.git
    GitClone ${BASE_DIR} ${REPO}
}

GitClone golang.org/x/net https://go.googlesource.com/net


GithubClone stretchr/testify
GithubClone beorn7/perks
GithubClone golang/protobuf
GithubClone matttproud/golang_protobuf_extensions
GithubClone prometheus/client_golang
GithubClone prometheus/client_model
GithubClone prometheus/common
GithubClone prometheus/procfs


#
#GitClone golang.org/x/text https://go.googlesource.com/text
#GitClone golang.org/x/crypto https://go.googlesource.com/crypto
#GithubClone Sirupsen/logrus
#GitClone gopkg.in/gemnasium/logrus-airbrake-hook.v2 git@github.com:gemnasium/logrus-airbrake-hook.git
#GitClone gopkg.in/airbrake/gobrake.v2 git@github.com:airbrake/gobrake.git
#GithubClone julienschmidt/httprouter
#