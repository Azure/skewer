#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset
set -x

TOOLMOD="${GITHUB_WORKSPACE}/.tools"

function deps() {
    mkdir -p "${TOOLMOD}"
    pushd "${TOOLMOD}"
    go mod init tools
    go get -tool github.com/axw/gocov/gocov@latest
    go get -tool github.com/AlekSi/gocov-xml@latest
    go get -tool github.com/wadey/gocovmerge@latest
    go get golang.org/x/tools@latest
    go mod tidy
    popd
}

function init() {
    export GOTOOLCHAIN=go1.25.0
    go env
}

function test() {
    # Run tests for root module
    echo "Running tests for v1 module..."
    go test -v -race -coverprofile=coverage-v1.out -covermode=atomic ./...
    
    # Run tests for v2 module
    echo "Running tests for v2 module..."
    cd v2
    go test -v -race -coverprofile=../coverage-v2.out -covermode=atomic ./...
    cd ..
    
    # Merge coverage files
    echo "Merging coverage files..."
    go -C "${TOOLMOD}" tool gocovmerge "${GITHUB_WORKSPACE}/coverage-v1.out" "${GITHUB_WORKSPACE}/coverage-v2.out" > coverage.out
    
    # Convert merged coverage to XML
    go -C "${TOOLMOD}" tool gocov convert "${GITHUB_WORKSPACE}/coverage.out" | go -C "${TOOLMOD}" tool gocov-xml > coverage.xml
    
    # Clean up intermediate files
    rm coverage-v1.out coverage-v2.out coverage.out
}

init
deps
test
