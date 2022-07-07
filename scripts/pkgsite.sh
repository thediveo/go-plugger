#!/bin/bash
set -e

if ! command -v pkgsite &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v pkgsite &>/dev/null; then
        go install golang.org/x/pkgsite/cmd/pkgsite@master
    fi
fi

NPMBIN=$(npm bin)
export PATH="$NPMBIN:$PATH"
if ! command -v browser-refresh &>/dev/null; then
    npm install browser-refresh
fi

if ! command -v nodemon &>/dev/null; then
    npm install nodemon
fi

# https://stackoverflow.com/a/2173421
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# https://mdaverde.com/posts/golang-local-docs
browser-sync start --port 6060 --proxy localhost:6061 --reload-delay 2000 --reload-debounce 5000 --no-ui --no-open &
PKGSITE=$(which pkgsite)
nodemon --signal SIGTERM --watch './**/*' -e go --exec "browser-sync --port 6060 reload && $PKGSITE -http=localhost:6061 ."
