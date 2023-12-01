make test:
	go test ./... -v -race

make benchTest:
	go test -bench=

make coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out