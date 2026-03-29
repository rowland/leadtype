.PHONY: binaries test samples ltml-samples ltml-samples-open ltml-image-sample-local ltml-image-sample-remote

SAMPLES := $(sort $(wildcard samples/*/*.go))
BIN_DIR := bin
BINARY_PKGS := ./cmd/render-ltml ./cmd/serve-ltml ./ttdump
ARABIC_BINARY_PKGS := ./cmd/render-ltml ./cmd/serve-ltml
LTML_IMAGE_SAMPLE := ltml/samples/test_031_render_ltml_images.ltml
LTML_IMAGE_JPEG := pdf/testdata/testimg.jpg
LTML_IMAGE_PNG := pdf/testdata/eidetic.png
LTML_IMAGE_LOCAL_OUTPUT := ltml/samples/test_031_render_ltml_images.local.pdf
LTML_IMAGE_REMOTE_OUTPUT := ltml/samples/test_031_render_ltml_images.remote.pdf
LTML_SERVER_ADDR ?= 127.0.0.1:18080

binaries:
	@mkdir -p $(BIN_DIR)
	@for pkg in $(ARABIC_BINARY_PKGS); do \
		name=$$(basename $$pkg); \
		echo "==> go build -tags arabic -o $(BIN_DIR)/$$name $$pkg"; \
		go build -tags arabic -o $(BIN_DIR)/$$name $$pkg || exit $$?; \
	done
	@for pkg in $(filter-out $(ARABIC_BINARY_PKGS),$(BINARY_PKGS)); do \
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

ltml-samples:
	go test -tags arabic ./ltml -run TestSamples -write-sample-pdfs

ltml-samples-open:
	go test -tags arabic ./ltml -run TestSamples -args -open-sample-pdfs

ltml-image-sample-local: binaries
	@echo "==> rendering $(LTML_IMAGE_SAMPLE) locally"
	@./$(BIN_DIR)/render-ltml \
		-e $(LTML_IMAGE_JPEG) \
		-e $(LTML_IMAGE_PNG) \
		-o $(LTML_IMAGE_LOCAL_OUTPUT) \
		$(LTML_IMAGE_SAMPLE)
	@echo "wrote $(LTML_IMAGE_LOCAL_OUTPUT)"

ltml-image-sample-remote: binaries
	@echo "==> rendering $(LTML_IMAGE_SAMPLE) through serve-ltml on $(LTML_SERVER_ADDR)"
	@set -e; \
	./$(BIN_DIR)/serve-ltml -listen $(LTML_SERVER_ADDR) -assets . >/tmp/leadtype-serve-ltml.log 2>&1 & \
	pid=$$!; \
	trap 'kill $$pid 2>/dev/null || true; wait $$pid 2>/dev/null || true' EXIT; \
	sleep 1; \
	./$(BIN_DIR)/render-ltml \
		-submit http://$(LTML_SERVER_ADDR)/render \
		-e $(LTML_IMAGE_JPEG) \
		-e $(LTML_IMAGE_PNG) \
		-o $(LTML_IMAGE_REMOTE_OUTPUT) \
		$(LTML_IMAGE_SAMPLE)
	@echo "wrote $(LTML_IMAGE_REMOTE_OUTPUT)"
