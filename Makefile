GOVERSION=$(shell go version)
GOOS=$(word 1,$(subst /, ,$(lastword $(GOVERSION))))
GOARCH=$(word 2,$(subst /, ,$(lastword $(GOVERSION))))
VERSION=0.1.1
RELEASE_DIR=releases
ARTIFACTS_DIR=$(RELEASE_DIR)/artifacts/$(VERSION)
GITHUB_USERNAME=syohex

.PHONY: clean build build-linux-amd64 $(RELEASE_DIR)/$(GOOS)/$(GOARCH)/byzanz-window

build-linux-amd64:
	@$(MAKE) build GOOS=linux GOARCH=amd64

$(RELEASE_DIR)/byzanz-window_$(GOOS)_$(GOARCH)/byzanz-window:
	go build -o $(RELEASE_DIR)/byzanz-window_$(GOOS)_$(GOARCH)/byzanz-window cmd/byzanz-window/byzanz-window.go

build: $(RELEASE_DIR)/byzanz-window_$(GOOS)_$(GOARCH)/byzanz-window

all: build-linux-amd64

release: release-linux-amd64

$(RELEASE_DIR)/byzanz-window_$(GOOS)_$(GOARCH)/README.md:
	@cp README.md $(RELEASE_DIR)/byzanz-window_$(GOOS)_$(GOARCH)

release-readme: $(RELEASE_DIR)/byzanz-window_$(GOOS)_$(GOARCH)/README.md

release-linux-amd64: build-linux-amd64
	@$(MAKE) release-readme release-targz GOOS=linux GOARCH=amd64

$(ARTIFACTS_DIR):
	@mkdir -p $(ARTIFACTS_DIR)

release-targz: $(ARTIFACTS_DIR)
	tar -czf $(ARTIFACTS_DIR)/byzanz-window_$(GOOS)_$(GOARCH).tar.gz -C $(RELEASE_DIR) byzanz-window_$(GOOS)_$(GOARCH)

release-github-token: github_token
	@echo "file `github_token` is required"

release-upload: release-linux-amd64
	ghr -u $(GITHUB_USERNAME) -t $(shell cat github_token) --draft --replace $(VERSION) $(ARTIFACTS_DIR)

clean:
	-rm $(RELEASE_DIR)/*/*
	-rm $(ARTIFACTS_DIR)/*
