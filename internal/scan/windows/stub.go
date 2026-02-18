//go:build !windows

package windows

type Neighbor struct {
	IP  string
	MAC string
}

func LookupMAC(ip string) string {
	return ""
}

func Neighbors() []Neighbor {
	return nil
}
