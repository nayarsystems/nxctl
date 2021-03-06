#!/bin/bash

set -x

GOOS=linux GOARCH=amd64 go build -o nxctl.linux.amd64
GOOS=linux GOARCH=386 go build -o nxctl.linux.386

GOOS=darwin GOARCH=amd64 go build -o nxctl.darwin.amd64
GOOS=darwin GOARCH=386 go build -o nxctl.darwin.386

GOOS=windows GOARCH=amd64 go build -o nxctl.windows.amd64.exe
GOOS=windows GOARCH=386 go build -o nxctl.windows.386.exe

GOARCH=arm go build -o nxctl.arm
