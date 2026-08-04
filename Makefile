BINARY_NAME=frizzlabs-ddns
BUILD_DIR=bin
CMD_DIR=./cmd/frizzlabs-ddns
VERSION ?= 1.0.0
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -s -w"

.PHONY: all build clean test bench lint fmt vet install uninstall

all: build test

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

test:
	go test -v -race -cover ./...

bench:
	go test -bench=. -benchmem ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping static analysis"; \
	fi

install: build
	install -d /usr/local/bin
	install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	install -d /etc/frizzlabs-ddns
	install -d /var/lib/frizzlabs-ddns
	@if [ ! -f /etc/frizzlabs-ddns/config.json ]; then \
		install -m 0600 configs/config.example.json /etc/frizzlabs-ddns/config.json; \
	fi
	install -m 0644 systemd/frizzlabs-ddns.service /etc/systemd/system/
	install -m 0644 systemd/frizzlabs-ddns.timer /etc/systemd/system/
	systemctl daemon-reload
	@echo "Installation complete. Edit /etc/frizzlabs-ddns/config.json and enable timer:"
	@echo "systemctl enable --now frizzlabs-ddns.timer"

uninstall:
	systemctl disable --now frizzlabs-ddns.timer 2>/dev/null || true
	rm -f /etc/systemd/system/frizzlabs-ddns.service
	rm -f /etc/systemd/system/frizzlabs-ddns.timer
	rm -f /usr/local/bin/$(BINARY_NAME)
	systemctl daemon-reload
	@echo "Uninstallation complete. Preserved /etc/frizzlabs-ddns and /var/lib/frizzlabs-ddns."
