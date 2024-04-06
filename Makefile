.PHONY: build clean install

build: mooltah.go
	go build mooltah.go

install: build
	cp mooltah ${HOME}/bin/


clean:
	rm -f mooltah