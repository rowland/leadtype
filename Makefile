.PHONY: test samples

SAMPLES := $(sort $(wildcard samples/*/*.go))

test:
	go test ./...

samples:
	@for sample in $(SAMPLES); do \
		echo "==> go run $$sample"; \
		go run $$sample || exit $$?; \
	done
