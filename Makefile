test:
	go test -v -coverprofile=cover.out ./...

bench:
	cd examples/benchmarks/ && sqlite3 boiler.db "CREATE TABLE IF NOT EXISTS records (id integer primary key, name text)"
	cd examples/benchmarks/ && sqlboiler sqlite3
	cd examples/benchmarks/ && go run *.go && rm -rf *db && rm -rf *journal