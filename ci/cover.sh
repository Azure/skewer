#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset
set -x

function install() {
    REPO="$1"
    APP="$2"
    mkdir -p "${GITHUB_WORKSPACE}/tmp/$APP"
    pushd "${GITHUB_WORKSPACE}/tmp/$APP"
    go mod init tmp
    go install "$REPO"
    cp "$(go env GOPATH)/bin/$APP" "${GITHUB_WORKSPACE}/bin/$APP"
    popd
    rm -rf "${GITHUB_WORKSPACE}/tmp/$APP"
    file "${GITHUB_WORKSPACE}/bin/$APP"
}

function deps() {
    install github.com/axw/gocov/... gocov
    install github.com/AlekSi/gocov-xml gocov-xml
}

function init() {
    go env
    mkdir -p "${GITHUB_WORKSPACE}/bin"
    mkdir -p "${GITHUB_WORKSPACE}/tmp"
    export PATH=$PATH:${GITHUB_WORKSPACE}/bin
}

function test() {
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    gocov convert coverage.out | gocov-xml > coverage.xml
    rm coverage.out
}

init
deps
test
