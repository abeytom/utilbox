# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=bin/utils

LINUX_OUT_BASE=./bin/linux-64/utilbox
BINARY_UNIX=$(LINUX_OUT_BASE)/go/bin/utils

OSX_OUT_BASE=./bin/osx/utilbox
BINARY_OSX=$(OSX_OUT_BASE)/go/bin/utils

all: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
	find ../ -maxdepth 1 -type f ! -name "*.*" -exec cp "{}" $(LINUX_OUT_BASE) \;
	tar -zcvf bin/utilbox-linux-amd64.tar.gz -C ./bin/linux-64/ .
build-osx:
	$(GOBUILD) -o $(BINARY_OSX) -v
	find ../ -maxdepth 1 -type f ! -name "*.*" -exec cp "{}" $(OSX_OUT_BASE) \;
	tar -zcvf bin/utilbox-osx.tar.gz -C ./bin/osx/ .

build-tar : build-linux build-osx

# git tag v0.2 -a && git push origin --tags



