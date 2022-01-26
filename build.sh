#!/bin/bash

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./DictGenerate_windows_amd64.exe ./DictGenerate.go

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./DictGenerate_darwin_amd64 ./DictGenerate.go
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./DictGenerate_darwin_arm64 ./DictGenerate.go

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./DictGenerate_linux_amd64 ./DictGenerate.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -o ./DictGenerate_linux_armv5 ./DictGenerate.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -o ./DictGenerate_linux_armv6 ./DictGenerate.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o ./DictGenerate_linux_armv7 ./DictGenerate.go