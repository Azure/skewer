#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset
set -x

function deps() {
    mkdir -p "${GITHUB_WORKSPACE}/tmp"
    pushd "${GITHUB_WORKSPACE}/tmp"
    go mod init tmp
    go install github.com/axw/gocov/gocov@latest
    go install github.com/AlekSi/gocov-xml@latest
    go install github.com/wadey/gocovmerge@latest
    cp "$(go env GOPATH)/bin/gocov" "${GITHUB_WORKSPACE}/bin/gocov"
    cp "$(go env GOPATH)/bin/gocov-xml" "${GITHUB_WORKSPACE}/bin/gocov-xml"
    cp "$(go env GOPATH)/bin/gocovmerge" "${GITHUB_WORKSPACE}/bin/gocovmerge"
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
    gocovmerge coverage-v1.out coverage-v2.out > coverage.out
    
    # Convert merged coverage to XML
    gocov convert coverage.out | gocov-xml > coverage.xml
    
    # Clean up intermediate files
    rm coverage-v1.out coverage-v2.out coverage.out
}

init
deps
test
