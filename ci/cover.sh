#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset
set -x

function deps() {
    mkdir -p "${GITHUB_WORKSPACE}/tmp"
    pushd "${GITHUB_WORKSPACE}/tmp"
    go mod init tmp
    go install github.com/axw/gocov/gocov
    go install github.com/AlekSi/gocov-xml
    cp "$(go env GOPATH)/bin/gocov" "${GITHUB_WORKSPACE}/bin/gocov"
    cp "$(go env GOPATH)/bin/gocov-xml" "${GITHUB_WORKSPACE}/bin/gocov-xml"
    popd
    rm -rf "${GITHUB_WORKSPACE}/tmp"
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
