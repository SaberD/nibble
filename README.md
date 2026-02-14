# Nibble
Local network scanner that focuses on speed and ease of use with hardware identification and a terminal UI.

Select a network interface, and Nibble scans your local subnet. Lists hosts, hardware manufacturer, open ports and their services.

![demo](demo.svg)


- **Hardware identification** — Maps each device MAC address to a likely vendor (for example, Raspberry Pi, Ubiquiti, Apple), so unknown IPs are easier to recognize
- **Banner grabbing** — Reads service banners on open ports to show what software is running (for example, OpenSSH or nginx versions), so you can identify services
- **Multi-port scanning** — SSH, Telnet, HTTP, HTTPS, SMB, RDP, and more
- **Smart interface filtering** — Skips loopback and irrelevant adapters

## Installation
go:
```bash
go install github.com/saberd/nibble@latest
```
brew:
```bash
brew install saberd/tap/nibble
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

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
