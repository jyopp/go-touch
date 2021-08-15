
GOTARGET=GOOS=linux GOARCH=arm GOARM=5
REMOTE=pylit-2.local:~/touchInput/

# Check the build, by building everything for ARM w/o saving output
check:
	$(GOTARGET) go build ./...

# Build the basic sample and install to the rmeote
install-basic:
	$(GOTARGET) go build -o "bin_arm/" ./examples/basic
	@scp -C "bin_arm/basic" $(REMOTE)

.PHONY: check
