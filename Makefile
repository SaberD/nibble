.PHONY: all build demo update run pip npm

all: run

build:
	@go build -o nibble .
	@echo "Built nibble binary"

nibble: build

run: nibble
	@./nibble

demo: nibble
	@TERM=xterm-256color asciinema rec demo.cast --overwrite -c "bash -i -c 'NIBBLE_DEMO=1 PS1=\"user@machine:~/nibble\$$ \" ./nibble'"
	@LAST=$$(tail -1 demo.cast | grep -oP '^\[\K[0-9.]+'); \
		END=$$(awk "BEGIN {print $$LAST + 3}"); \
		echo "[$$END, \"o\", \"\"]" >> demo.cast
	@SVG_HEIGHT=$$(head -n 1 demo.cast | grep -oP '"height":\s*\K[0-9]+'); \
	svg-term --in demo.cast --out demo.svg --window --height "$$SVG_HEIGHT"
	@echo "Generated demo.svg"

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
