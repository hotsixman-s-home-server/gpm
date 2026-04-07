GOOS=linux GOARCH=amd64 go build -o dist/linux.bin
GOOS=darwin GOARCH=arm64 go build -o dist/mac-arm.bin
GOOS=windows GOARCH=amd64 go build -o dist/windows-x64.exe
GOOS=windows GOARCH=386 go build -o dist/windows-x32.exe