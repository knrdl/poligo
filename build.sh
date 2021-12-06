#!/bin/bash

docker run -it --rm -v "$PWD:$PWD" -w "$PWD" golang go build -ldflags="-s -w"
./poligo --timeout=500ms cwd-exists warn-memory=75% term-title current-time go-version python-version nodejs-project docker-version kernel-version warn-offline shell-level virtual-env work-dir=4 sudo-root git read-only ssh-connection user-name=your_default_username exit-code=$?
