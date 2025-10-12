all:
	go run main.go

check: 
	gofmt -w -s .
	go vet .
	CGO_ENABLED=0 staticcheck
