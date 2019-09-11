#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./DictGenerate_linux_amd64 ./DictGenerate.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./DictGenerate_darwin_amd64 ./DictGenerate.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./DictGenerate_windows_amd64.exe ./DictGenerate.go
