//go:build !darwin

package macos

type Neighbor struct {
	IP    string
	MAC   string
	Iface string
}

func LookupMAC(ip string) string {
	return ""
}

func Neighbors(ifaceName string) []Neighbor {
	return nil
}
