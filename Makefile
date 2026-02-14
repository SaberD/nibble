.PHONY: all build demo update run

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
	@svg-term --in demo.cast --out demo.svg --window
	@echo "Generated demo.svg"

update:
	@echo "Downloading IEEE OUI database..."
	@curl -sL "https://standards-oui.ieee.org/oui/oui.csv" -o internal/scan/oui.csv
	@echo "Updated $$(wc -l < internal/scan/oui.csv) entries"




