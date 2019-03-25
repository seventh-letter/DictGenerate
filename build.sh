#!/bin/bash

go build DictGenerate.go
mv DictGenerate DictGenerate_darwin_amd64

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build DictGenerate.go
mv DictGenerate DictGenerate_linux_amd64

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build DictGenerate.go
mv DictGenerate.exe DictGenerate_windows_amd64.exe