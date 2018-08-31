#! /usr/bin/env bash

PLUGIN_PATH=$GOPATH/src/github.com/mcb/cf-accounting-report/cmd/accounting-report
PLUGIN_NAME=$(basename $PLUGIN_PATH)

GOOS=linux GOARCH=amd64 go build -o ${PLUGIN_NAME}.linux64 cmd/${PLUGIN_NAME}/${PLUGIN_NAME}.go
GOOS=linux GOARCH=386 go build -o ${PLUGIN_NAME}.linux32 cmd/${PLUGIN_NAME}/${PLUGIN_NAME}.go
GOOS=windows GOARCH=amd64 go build -o ${PLUGIN_NAME}.win64 cmd/${PLUGIN_NAME}/${PLUGIN_NAME}.go
GOOS=windows GOARCH=386 go build -o ${PLUGIN_NAME}.win32 cmd/${PLUGIN_NAME}/${PLUGIN_NAME}.go
GOOS=darwin GOARCH=amd64 go build -o ${PLUGIN_NAME}.osx cmd/${PLUGIN_NAME}/${PLUGIN_NAME}.go

shasum -a 1 ${PLUGIN_NAME}.* > buildinfo.txt