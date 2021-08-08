
GONAME=$(shell basename "$(PWD)")
GOTARGET=GOOS=linux GOARCH=arm GOARM=5
REMOTE=pylit-2.local:~/touchInput/

install:
	$(GOTARGET) go build .
	@scp -r $(GONAME) $(REMOTE)

.PHONY: install
