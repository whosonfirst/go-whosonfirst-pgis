cli-tools:
	go build -mod vendor -o bin/wof-pgis-connect cmd/wof-pgis-connect/main.go
	go build -mod vendor -o bin/wof-pgis-dump cmd/wof-pgis-dump/main.go
	go build -mod vendor -o bin/wof-pgis-index cmd/wof-pgis-index/main.go
	go build -mod vendor -o bin/wof-pgis-prune cmd/wof-pgis-prune/main.go
