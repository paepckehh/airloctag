# generate internal/airports/airports.go package

all:
	go run -mod=readonly generate.go

update:
	curl -O https://raw.githubusercontent.com/datasets/airport-codes/refs/heads/main/data/airport-codes.csv
	sed -i '1d' airport-codes.csv
	go run -mod=readonly generate.go
