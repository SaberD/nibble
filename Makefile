.PHONY: all build demo demo-gif update run pip npm goreleaser

all: run

build:
	@go build -o nibble .
	@echo "Built nibble binary"

nibble: build

run: nibble
	@./nibble

demo: demo-gif

demo-gif: nibble
	@if ! command -v vhs >/dev/null 2>&1; then \
		echo "vhs not found. Install it from https://github.com/charmbracelet/vhs"; \
		exit 1; \
	fi
	@vhs demo.tape
	@echo "Generated demo.gif"

update:
	@echo "Downloading IEEE OUI database..."
	@curl -sL "https://standards-oui.ieee.org/oui/oui.csv" -o internal/scan/oui.csv
	@echo "Updated $$(wc -l < internal/scan/oui.csv) entries"

pip:
	@cd python-package && \
	if python3 -m venv .venv >/dev/null 2>&1; then \
		. .venv/bin/activate && \
		python -m pip install -U pip build twine && \
		python -m build && \
		python -m twine check dist/*; \
	else \
		echo "python3-venv not available; using user-site fallback"; \
		python3 -m pip install --user --break-system-packages -U build twine && \
		python3 -m build && \
		python3 -m twine check dist/*; \
	fi
	@echo "Built Python package in python-package/dist"

npm:
	@cd npm-package && npm pack --silent
	@echo "Built npm package tarball in npm-package/"

goreleaser:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Installing goreleaser locally..."; \
		TMP_DIR=$$(mktemp -d); \
		curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar xz -C "$$TMP_DIR"; \
		sudo mv "$$TMP_DIR/goreleaser" /usr/local/bin/; \
		rm -rf "$$TMP_DIR"; \
	fi
	@goreleaser check
	@goreleaser release --snapshot --clean
	@echo "GoReleaser snapshot validation passed"
