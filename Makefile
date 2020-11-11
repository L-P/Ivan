EXEC=./$(shell basename "$(shell pwd)")
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "unknown")
RELEASE_DIR=ivan_release_$(VERSION)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'

all: $(EXEC) tags

$(EXEC):
	go build $(BUILDFLAGS)

.PHONY: $(EXEC) vendor upgrade lint test coverage debian-deps release clean tags

tags:
	ctags-universal -R timer tracker input-viewer *.go

clean:
	rm -rf ivan_release_*
	rm -f ivan ivan.exe

release:
	rm -rf "$(RELEASE_DIR)"
	mkdir -p "$(RELEASE_DIR)/assets"
	cp assets/*.png assets/*.json assets/LICENSE.md "$(RELEASE_DIR)/assets"
	pandoc -M "pagetitle=Ivan Item Tracker" -s -o "$(RELEASE_DIR)/readme.html" < README.md

	GOOS="linux" GOARCH="amd64" make $(EXEC)
	mv "$(EXEC)" "$(RELEASE_DIR)/ivan.linux64"
	tar czf "$(RELEASE_DIR)_linux64.tgz" "$(RELEASE_DIR)"
	rm "$(RELEASE_DIR)/ivan.linux64"

	GOOS="windows" GOARCH="amd64" make $(EXEC)
	mv "$(EXEC).exe" "$(RELEASE_DIR)/ivan.exe"
	zip -r "$(RELEASE_DIR)_window64.zip" "$(RELEASE_DIR)"

debian-deps:
	# Ebiten
	sudo apt install libc6-dev libglu1-mesa-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config

coverage:
	go test -tags docker,api -covermode=count -coverprofile=coverage.cov --timeout=30s ./...
	go tool cover -html=coverage.cov -o coverage.html
	rm coverage.cov
	sensible-browser coverage.html

test:
	go test ./...

vendor:
	go get -v
	go mod vendor
	go mod tidy

upgrade:
	go get -u -v
	go mod vendor
	go mod tidy

lint: $(GOLANGCI)
	golangci-lint run
