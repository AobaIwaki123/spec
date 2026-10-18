.PHONY: all build install test vet verify clean

BINARY_NAME := spec
INSTALL_DIR := $(HOME)/.local/bin

all: build

build:
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/spec

install:
	mkdir -p $(INSTALL_DIR)
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	go build -o $(INSTALL_DIR)/$(BINARY_NAME) ./cmd/spec
	ln -sf $(INSTALL_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/db-quiz
	@echo "Successfully installed $(BINARY_NAME) to $(INSTALL_DIR)/$(BINARY_NAME)"

test:
	go test -v ./...

vet:
	go vet ./...

verify: vet test build

clean:
	rm -rf bin/
