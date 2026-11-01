BINARY := dronagiri
PREFIX ?= $(HOME)/.local

.PHONY: build install uninstall test fmt vet check logo clean

build:
	go build -o $(BINARY) .

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "installed to $(PREFIX)/bin/$(BINARY)"

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test

logo:
	go run ./cmd/logo-svg

clean:
	rm -f $(BINARY)
