GOPATH=$(shell pwd)/vendor:$(shell pwd)
GOBIN=$(shell pwd)/
GONAME=$(shell basename "$(PWD)")
GOTARGET=GOOS=linux GOARCH=arm GOARM=5
REMOTE=pylit-2.local:~/touchInput/

build:
	@echo "Building $(GONAME)"
	$(GOTARGET) GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build .

get:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get .

install:
	$(GOTARGET) GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build .
	@scp -r $(GONAME) $(REMOTE)

clear:
	@clear

clean:
	@echo "Cleaning"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

.PHONY: build get install run watch start stop restart clean
