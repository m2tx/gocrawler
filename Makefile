test:
	mkdir -p coverage
	go test ./... -cover -coverprofile=./coverage/coverage.out
	go tool cover -html=./coverage/coverage.out