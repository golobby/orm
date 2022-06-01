test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out

bench:
	cd benchmark/ && go test -v -bench=. -benchmem