package portsview

import "github.com/backendsystems/nibble/internal/scanner"

type Model struct {
	ShowHelp      bool
	PortPack      string
	CustomPorts   string
	CustomCursor  int
	PortConfigLoc string
	ErrorMsg      string
	NetworkScan   scanner.Scanner
}
