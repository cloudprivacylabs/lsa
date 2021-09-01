build:
	go generate ../...
	go build

all:
	env GOOS=linux GOARCH=amd64 go build -o layers.linux-amd64
	env GOOS=windows GOARCH=amd64 go build -o layers.windows-amd64
	env GOOS=darwin GOARCH=amd64 go build -o layers.darwin-amd64
