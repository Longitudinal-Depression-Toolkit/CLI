BINARY ?= ldt
INSTALL_DIR ?= $(HOME)/.local/bin
BASH_RC ?= $(HOME)/.bashrc
FISH_CONFIG ?= $(HOME)/.config/fish/config.fish

.PHONY: build install-bash install-fish fmt fmt-check vet lint lint-install check

build:
	go build -o $(BINARY) .

install-bash: build
	@mkdir -p "$(INSTALL_DIR)"
	@cp "$(BINARY)" "$(INSTALL_DIR)/$(BINARY)"
	@chmod +x "$(INSTALL_DIR)/$(BINARY)"
	@touch "$(BASH_RC)"
	@if ! grep -Fq 'export PATH="$(INSTALL_DIR):$$PATH"' "$(BASH_RC)"; then \
		echo 'export PATH="$(INSTALL_DIR):$$PATH"' >> "$(BASH_RC)"; \
		echo "Added PATH export to $(BASH_RC)"; \
	fi
	@echo "Installed $(BINARY) to $(INSTALL_DIR)/$(BINARY)"
	@echo "Run: source $(BASH_RC)"

install-fish: build
	@mkdir -p "$(INSTALL_DIR)"
	@cp "$(BINARY)" "$(INSTALL_DIR)/$(BINARY)"
	@chmod +x "$(INSTALL_DIR)/$(BINARY)"
	@mkdir -p "$$(dirname "$(FISH_CONFIG)")"
	@touch "$(FISH_CONFIG)"
	@if ! grep -Fq 'fish_add_path $(INSTALL_DIR)' "$(FISH_CONFIG)"; then \
		echo 'fish_add_path $(INSTALL_DIR)' >> "$(FISH_CONFIG)"; \
		echo "Added fish PATH setup to $(FISH_CONFIG)"; \
	fi
	@echo "Installed $(BINARY) to $(INSTALL_DIR)/$(BINARY)"
	@echo "Run: source $(FISH_CONFIG)"

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

fmt-check:
	@files=$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*')); \
	if [ -n "$$files" ]; then \
		echo "Go files need formatting:"; \
		echo "$$files"; \
		exit 1; \
	fi

vet:
	go vet ./...

lint:
	golangci-lint run ./...

lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

check: fmt-check vet lint
