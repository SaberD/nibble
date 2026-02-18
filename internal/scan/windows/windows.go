//go:build windows

package windows

import (
	"encoding/binary"
	"net"
	"net/netip"
	"unsafe"

	"github.com/backendsystems/nibble/internal/scan/shared"
	syswin "golang.org/x/sys/windows"
)

const maxPhysAddrLen = 8

type Neighbor struct {
	IP  string
	MAC string
}

type mibIPNetRow struct {
	Index       uint32
	PhysAddrLen uint32
	PhysAddr    [maxPhysAddrLen]byte
	Addr        uint32
	Type        uint32
}

var (
	modiphlpapi    = syswin.NewLazySystemDLL("iphlpapi.dll")
	procGetIPTable = modiphlpapi.NewProc("GetIpNetTable")
)

func LookupMAC(ip string) string {
	for _, row := range Neighbors() {
		if row.IP == ip && row.MAC != "00:00:00:00:00:00" {
			return row.MAC
		}
	}
	return ""
}

func Neighbors() []Neighbor {
	rows, err := readIPNetTable()
	if err != nil {
		return nil
	}

	out := make([]Neighbor, 0, len(rows))
	for _, row := range rows {
		if row.PhysAddrLen == 0 || row.PhysAddrLen > maxPhysAddrLen {
			continue
		}

		ipBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(ipBytes, row.Addr)
		ip, ok := netip.AddrFromSlice(ipBytes)
		if !ok || !ip.Is4() {
			continue
		}

		mac := shared.NormalizeMAC(rawMAC(row.PhysAddr[:row.PhysAddrLen]))
		if mac == "" {
			continue
		}

		out = append(out, Neighbor{IP: ip.String(), MAC: mac})
	}

	return out
}

func readIPNetTable() ([]mibIPNetRow, error) {
	var size uint32
	r0, _, _ := procGetIPTable.Call(0, uintptr(unsafe.Pointer(&size)), 0)
	if r0 != uintptr(syswin.ERROR_INSUFFICIENT_BUFFER) {
		if r0 == 0 || size == 0 {
			return nil, nil
		}
		return nil, syswin.Errno(r0)
	}

	buf := make([]byte, size)
	r1, _, _ := procGetIPTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0)
	if r1 != 0 {
		return nil, syswin.Errno(r1)
	}

	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	if count == 0 {
		return nil, nil
	}

	headerSize := unsafe.Sizeof(count)
	entrySize := unsafe.Sizeof(mibIPNetRow{})
	need := headerSize + uintptr(count)*entrySize
	if need > uintptr(len(buf)) {
		return nil, nil
	}

	rowsPtr := (*mibIPNetRow)(unsafe.Add(unsafe.Pointer(&buf[0]), headerSize))
	rowsView := unsafe.Slice(rowsPtr, int(count))
	rows := make([]mibIPNetRow, len(rowsView))
	copy(rows, rowsView)

	return rows, nil
}

func rawMAC(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	hw := make([]byte, 6)
	copy(hw, b[:6])
	return net.HardwareAddr(hw).String()
}
