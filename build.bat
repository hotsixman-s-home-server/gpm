set GOOS=linux
set GOARCH=amd64
go build -o dist/linux.bin

set GOOS=darwin
set GOARCH=arm64
go build -o dist/mac-arm.bin

set GOOS=windows
set GOARCH=amd64
go build -o dist/windows-x64.exe

set GOOS=windows
set GOARCH=386
go build -o dist/windows-x32.exe