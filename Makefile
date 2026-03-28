.PHONY: binaries test samples

SAMPLES := $(sort $(wildcard samples/*/*.go))
BIN_DIR := bin
BINARY_PKGS := ./cmd/render-ltml ./cmd/serve-ltml ./ttdump

binaries:
	@mkdir -p $(BIN_DIR)
	@for pkg in $(BINARY_PKGS); do \
		name=$$(basename $$pkg); \
		echo "==> go build -o $(BIN_DIR)/$$name $$pkg"; \
		go build -o $(BIN_DIR)/$$name $$pkg || exit $$?; \
	done

test:
	go test ./...

samples:
	@for sample in $(SAMPLES); do \
		echo "==> go run $$sample"; \
		go run $$sample || exit $$?; \
	done
