.PHONY: all build demo demo-cast upload update run pip npm goreleaser

all: run

build:
	@go build -o nibble .
	@echo "Built nibble binary"

nibble: build

run: nibble
	@./nibble

demo: demo-cast

demo-cast: nibble
	@if ! command -v asciinema >/dev/null 2>&1; then \
		echo "asciinema not found. Install it with: sudo apt install asciinema"; \
		exit 1; \
	fi
	@TERM=xterm-256color asciinema rec demo.cast --overwrite -c "bash -i -c 'NIBBLE_DEMO=1 PS1=\"user@machine:~/nibble\$$ \" ./nibble'"
	@LAST=$$(tail -1 demo.cast | grep -oP '^\[\K[0-9.]+'); \
		END=$$(awk "BEGIN {print $$LAST + 3}"); \
		echo "[$$END, \"o\", \"\"]" >> demo.cast
	@echo "Generated demo.cast"

upload:
	@if [ ! -f demo.cast ]; then \
		echo "demo.cast not found. Run 'make demo' first."; \
		exit 1; \
	fi
	@if ! command -v asciinema >/dev/null 2>&1; then \
		echo "asciinema not found. Install it with: sudo apt install asciinema"; \
		exit 1; \
	fi
	@asciinema upload demo.cast

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
