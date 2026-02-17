# Nibble
Nibble is a CLI tool for local network scanning that focuses on speed and ease of use.

Select a network interface, and Nibble scans your local subnet. Lists hosts, hardware manufacturer, open ports and their services.

![Nibble demo](demo.gif "Made with Bubble Tea VHS")


- **Hardware identification** — Maps each device MAC address to a likely vendor (for example, Raspberry Pi, Ubiquiti, Apple), so unknown IPs are easier to recognize
- **Banner grabbing** — Reads service banners on open ports to show what software is running (for example, OpenSSH or nginx versions), so you can identify services
- **Multi-port scanning** — SSH, Telnet, HTTP, HTTPS, SMB, RDP, and more
- **Two-phase discovery** — First shows currently visible neighbors from the local ARP/neighbor table, then runs a full subnet sweep and skips already found hosts
- **Smart interface filtering** — Skips loopback and irrelevant adapters

## Hotkeys
`↑/↓/←/→`, `w/s/a/d`, `h/j/k/l`: selection  
`Enter`: confirm.  
`q` or `Ctrl+C`: quit.  
`?`: help.

## Installation
you may have to restart terminal to run `nibble` after install.


go:
```bash
go install github.com/backendsystems/nibble@latest
```
brew:
```bash
brew install backendsystems/tap/nibble
```
pip:
```bash
pipx install nibble-cli
```
npm:
```bash
npm install -g @backendsystems/nibble
```
or run without install
```bash
npx @backendsystems/nibble
```

## Usage
Run the CLI with `nibble`, select a network interface.  
Interface icons: `🔌` = Ethernet, `📶` = Wi-Fi, `📦` = Container, `🔒` = VPN.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
