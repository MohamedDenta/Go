// +build windows

package net

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTCPTable = modiphlpapi.NewProc("GetExtendedTcpTable")
)

var netConnectionKindMap = map[string][]netConnectionKindType{
	"tcp4": {kindTCP4},
}

func IOCounters() ([]IOCountersStat, error) {
	return IOCountersWithContext(context.Background())
}

func IOCountersWithContext(ctx context.Context) ([]IOCountersStat, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var ret []IOCountersStat

	for _, ifi := range ifs {
		// fmt.Println("network name:", ifi.Name)
		// if strings.HasPrefix(strings.ToLower(ifi.Name), "lo") {
		// 	continue
		// }
		c := IOCountersStat{
			Name: ifi.Name,
		}

		row := windows.MibIfRow{Index: uint32(ifi.Index)}
		e := windows.GetIfEntry(&row)
		if e != nil {
			return nil, os.NewSyscallError("GetIfEntry", e)
		}
		c.BytesSent = uint64(row.OutOctets)
		c.BytesRecv = uint64(row.InOctets)
		c.PacketsSent = uint64(row.OutUcastPkts)
		c.PacketsRecv = uint64(row.InUcastPkts)

		ret = append(ret, c)
	}
	return ret, nil
}

// Return a list of network connections
// Available kind:
//   reference to netConnectionKindMap
func Connections(pid int32) ([]ConnectionStat, error) {
	return ConnectionsPid(pid)
}

// ConnectionsPid Return a list of network connections opened by a process
func ConnectionsPid(pid int32) ([]ConnectionStat, error) {
	return ConnectionsPidWithContext(pid)
}

func ConnectionsPidWithContext(pid int32) ([]ConnectionStat, error) {
	return getProcInet(netConnectionKindMap["tcp4"], pid)
}

func getProcInet(kinds []netConnectionKindType, pid int32) ([]ConnectionStat, error) {
	stats := make([]ConnectionStat, 0)

	for _, kind := range kinds {
		s, err := getNetStatWithKind(kind)
		if err != nil {
			continue
		}

		if pid == 0 {
			stats = append(stats, s...)
		} else {
			for _, ns := range s {
				if ns.Pid != pid {
					continue
				}
				stats = append(stats, ns)
			}
		}
	}

	return stats, nil
}

func getNetStatWithKind(kindType netConnectionKindType) ([]ConnectionStat, error) {
	if kindType.filename == "" {
		return nil, fmt.Errorf("kind filename must be required")
	}

	switch kindType.filename {
	case kindTCP4.filename:
		return getTCPConnections(kindTCP4.family)
	}

	return nil, fmt.Errorf("invalid kind filename, %s", kindType.filename)
}

func getTableInfo(filename string, table interface{}) (index, step, length int) {
	switch filename {
	case kindTCP4.filename:
		index = int(unsafe.Sizeof(table.(pmibTCPTableOwnerPidAll).DwNumEntries))
		step = int(unsafe.Sizeof(table.(pmibTCPTableOwnerPidAll).Table))
		length = int(table.(pmibTCPTableOwnerPidAll).DwNumEntries)
	}

	return
}

func getTCPConnections(family uint32) ([]ConnectionStat, error) {
	var (
		p    uintptr
		buf  []byte
		size uint32

		pmibTCPTable pmibTCPTableOwnerPidAll
	)

	if family == 0 {
		return nil, fmt.Errorf("faimly must be required")
	}

	for {
		switch family {
		case kindTCP4.family:
			if len(buf) > 0 {
				pmibTCPTable = (*mibTCPTableOwnerPid)(unsafe.Pointer(&buf[0]))
				p = uintptr(unsafe.Pointer(pmibTCPTable))
			} else {
				p = uintptr(unsafe.Pointer(pmibTCPTable))
			}
		}

		err := getExtendedTcpTable(p,
			&size,
			true,
			family,
			tcpTableOwnerPidAll,
			0)
		if err == nil {
			break
		}
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return nil, err
		}
		buf = make([]byte, size)
	}

	var (
		index, step int
		length      int
	)

	stats := make([]ConnectionStat, 0)
	switch family {
	case kindTCP4.family:
		index, step, length = getTableInfo(kindTCP4.filename, pmibTCPTable)
	}

	if length == 0 {
		return nil, nil
	}

	for i := 0; i < length; i++ {
		switch family {
		case kindTCP4.family:
			mibs := (*mibTCPRowOwnerPid)(unsafe.Pointer(&buf[index]))
			ns := mibs.convertToConnectionStat()
			stats = append(stats, ns)
		}

		index += step
	}
	return stats, nil
}

// tcpStatuses https://msdn.microsoft.com/en-us/library/windows/desktop/bb485761(v=vs.85).aspx
var tcpStatuses = map[mibTCPState]string{
	1:  "CLOSED",
	2:  "LISTEN",
	3:  "SYN_SENT",
	4:  "SYN_RECEIVED",
	5:  "ESTABLISHED",
	6:  "FIN_WAIT_1",
	7:  "FIN_WAIT_2",
	8:  "CLOSE_WAIT",
	9:  "CLOSING",
	10: "LAST_ACK",
	11: "TIME_WAIT",
	12: "DELETE",
}

func getExtendedTcpTable(pTcpTable uintptr, pdwSize *uint32, bOrder bool, ulAf uint32, tableClass tcpTableClass, reserved uint32) (errcode error) {
	r1, _, _ := syscall.Syscall6(procGetExtendedTCPTable.Addr(), 6, pTcpTable, uintptr(unsafe.Pointer(pdwSize)), getUintptrFromBool(bOrder), uintptr(ulAf), uintptr(tableClass), uintptr(reserved))
	if r1 != 0 {
		errcode = syscall.Errno(r1)
	}
	return
}

func getUintptrFromBool(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

const anySize = 1

// type MIB_TCP_STATE int32
type mibTCPState int32

type tcpTableClass int32

const (
	tcpTableBasicListener tcpTableClass = iota
	tcpTableBasicConnections
	tcpTableBasicAll
	tcpTableOwnerPidListener
	tcpTableOwnerPidConnections
	tcpTableOwnerPidAll
	tcpTableOwnerModuleListener
	tcpTableOwnerModuleConnections
	tcpTableOwnerModuleAll
)

// TCP

type mibTCPRowOwnerPid struct {
	DwState      uint32
	DwLocalAddr  uint32
	DwLocalPort  uint32
	DwRemoteAddr uint32
	DwRemotePort uint32
	DwOwningPid  uint32
}

func (m *mibTCPRowOwnerPid) convertToConnectionStat() ConnectionStat {
	ns := ConnectionStat{
		Family: kindTCP4.family,
		Type:   kindTCP4.sockType,
		Laddr: Addr{
			IP:   parseIPv4HexString(m.DwLocalAddr),
			Port: uint32(decodePort(m.DwLocalPort)),
		},
		Raddr: Addr{
			IP:   parseIPv4HexString(m.DwRemoteAddr),
			Port: uint32(decodePort(m.DwRemotePort)),
		},
		Pid:    int32(m.DwOwningPid),
		Status: tcpStatuses[mibTCPState(m.DwState)],
	}

	return ns
}

type mibTCPTableOwnerPid struct {
	DwNumEntries uint32
	Table        [anySize]mibTCPRowOwnerPid
}

type pmibTCPTableOwnerPidAll *mibTCPTableOwnerPid

func decodePort(port uint32) uint16 {
	return syscall.Ntohs(uint16(port))
}

func parseIPv4HexString(addr uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", addr&255, addr>>8&255, addr>>16&255, addr>>24&255)
}
