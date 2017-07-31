CWD=$(shell pwd)
GOPATH := $(CWD)

build:	rmdeps deps fmt bin

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-pgis; then rm -rf src/github.com/whosonfirst/go-whosonfirst-pgis; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pgis
	cp -r index src/github.com/whosonfirst/go-whosonfirst-pgis/index
	cp -r vendor/src/* src/
	cp -r vendor/src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/src/github.com/tidwall src/github.com/

rmdeps:
	if test -d src; then rm -rf src; fi 

deps:   
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/pretty"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-crawl"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-geojson-v2"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-hash"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-placetypes"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"
	@GOPATH=$(GOPATH) go get -u "github.com/lib/pq"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt index/*.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-pgis-connect cmd/wof-pgis-connect.go
	@GOPATH=$(GOPATH) go build -o bin/wof-pgis-dump cmd/wof-pgis-dump.go
	@GOPATH=$(GOPATH) go build -o bin/wof-pgis-index cmd/wof-pgis-index.go
	@GOPATH=$(GOPATH) go build -o bin/wof-pgis-prune cmd/wof-pgis-prune.go
