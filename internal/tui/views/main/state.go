package mainview

import "net"

type Model struct {
	Interfaces   []net.Interface
	InterfaceMap map[string][]net.Addr
	Cursor       int
	ShowHelp     bool
	ErrorMsg     string
}
