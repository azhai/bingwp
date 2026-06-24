APP       = bwp
BIN_DIR   = bin
COMMANDS  := $(notdir $(patsubst %/.,%,$(wildcard cmd/*/.)))

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
RELEASE    = -ldflags "-s -w -X main.Version=$(VERSION)"
GOBUILD    = go build $(RELEASE)

.PHONY: one all build run clean

one:
	@echo "Build $(APP) (local) ..."
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) -o $(BIN_DIR)/$(APP) ./
	@for cmd in $(COMMANDS); do \
		CGO_ENABLED=0 $(GOBUILD) -o $(BIN_DIR)/$$cmd ./cmd/$$cmd; \
	done

all: clean one build

build:
	@echo "Cross-compiling ..."
	mkdir -p $(BIN_DIR)
	@for target in $(APP) $(COMMANDS); do \
		if [ "$$target" = "$(APP)" ]; then src=.; else src=./cmd/$$target; fi; \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64  $(GOBUILD) -o $(BIN_DIR)/$$target-$(VERSION).darwin-arm64  $$src && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64  $(GOBUILD) -o $(BIN_DIR)/$$target-$(VERSION).darwin-amd64  $$src && \
		CGO_ENABLED=0 GOOS=linux   GOARCH=arm64  $(GOBUILD) -o $(BIN_DIR)/$$target-$(VERSION).linux-arm64   $$src && \
		CGO_ENABLED=0 GOOS=linux   GOARCH=amd64  $(GOBUILD) -o $(BIN_DIR)/$$target-$(VERSION).linux-amd64   $$src && \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64  $(GOBUILD) -o $(BIN_DIR)/$$target-$(VERSION).windows-amd64.exe $$src; \
	done
	@echo "✅ Build success."

run:
	go run ./

clean:
	rm -rf $(BIN_DIR)
	@echo "✅ Clean complete."
