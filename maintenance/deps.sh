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

GitClone gopkg.in/yaml.v2 https://github.com/go-yaml/yaml.git
