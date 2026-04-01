#!/usr/bin/env bash
set -o errexit
set -o pipefail
set -o nounset
set -x

function deps() {
    mkdir -p "${GITHUB_WORKSPACE}/tmp"
    pushd "${GITHUB_WORKSPACE}/tmp"
    go mod init tmp
    cat > tools.go <<'EOF'
//go:build tools

package tools

import (
	_ "github.com/axw/gocov/gocov"
	_ "github.com/AlekSi/gocov-xml"
	_ "github.com/wadey/gocovmerge"
)
EOF
    go get github.com/axw/gocov/gocov@latest
    go get github.com/AlekSi/gocov-xml@latest
    go get github.com/wadey/gocovmerge@latest
    go get golang.org/x/tools@latest
    go mod tidy
    go install github.com/axw/gocov/gocov
    go install github.com/AlekSi/gocov-xml
    go install github.com/wadey/gocovmerge
    cp "$(go env GOPATH)/bin/gocov" "${GITHUB_WORKSPACE}/bin/gocov"
    cp "$(go env GOPATH)/bin/gocov-xml" "${GITHUB_WORKSPACE}/bin/gocov-xml"
    cp "$(go env GOPATH)/bin/gocovmerge" "${GITHUB_WORKSPACE}/bin/gocovmerge"
    popd
    rm -rf "${GITHUB_WORKSPACE}/tmp"
}

function init() {
    export GOTOOLCHAIN=go1.25.0
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
