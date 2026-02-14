# Nibble

Fast local network scanner with hardware identification and a terminal UI.

![demo](demo.svg)


- **Hardware identification** — Resolves MAC addresses to manufacturers via the IEEE OUI database
- **Banner grabbing** — Detects service versions (SSH, HTTP Server, etc.)
- **Multi-port scanning** — SSH, Telnet, HTTP, HTTPS, SMB, RDP, and more
- **Concurrent scanning** — 100 parallel goroutines, rate-limited for reliability
- **Smart interface filtering** — Skips loopback and irrelevant adapters
- **Interactive UI** — Select interfaces with arrow keys, live progress bar

## Installation
go:
```bash
go install github.com/saberd/nibble@latest
```
pip:
```bash
pip install nibble
```
npm:
```bash
npm install -g @saberd/nibble
```

Or run without installing:

```bash
npx @saberd/nibble
```

## Usage

Select a network interface, and Nibble scans your local subnet — discovering hosts, open ports, service banners, and hardware manufacturers.

type `?` for help in terminal

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [mdlayher/arp](https://github.com/mdlayher/arp) — ARP resolution
