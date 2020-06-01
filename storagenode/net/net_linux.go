// +build linux

package net

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"../utils"
)

// IOCounters returnes network I/O statistics for wlan and ethernet network
// interface installed on the system.
func IOCounters() ([]IOCountersStat, error) {
	lines, err := utils.ReadLines("/proc/net/dev")
	if err != nil {
		return nil, err
	}

	parts := make([]string, 2)

	statlen := len(lines) - 1

	ret := make([]IOCountersStat, 0, statlen)

	for _, line := range lines[2:] {
		separatorPos := strings.LastIndex(line, ":")
		if separatorPos == -1 {
			continue
		}
		parts[0] = line[0:separatorPos]
		parts[1] = line[separatorPos+1:]

		interfaceName := strings.TrimSpace(parts[0])
		if interfaceName == "" || strings.HasPrefix(interfaceName, "lo") {
			continue
		}

		fields := strings.Fields(strings.TrimSpace(parts[1]))
		bytesRecv, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return ret, err
		}
		packetsRecv, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return ret, err
		}
		bytesSent, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			return ret, err
		}
		packetsSent, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return ret, err
		}

		nic := IOCountersStat{
			Name:        interfaceName,
			BytesRecv:   bytesRecv,
			PacketsRecv: packetsRecv,
			BytesSent:   bytesSent,
			PacketsSent: packetsSent,
		}
		ret = append(ret, nic)
	}

	return ret, nil
}

var TCPStatuses = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}

type inodeMap struct {
	pid int32
	fd  uint32
}

type connTmp struct {
	fd       uint32
	family   uint32
	sockType uint32
	laddr    Addr
	raddr    Addr
	status   string
	pid      int32
	boundPid int32
	path     string
}

// Connections Return a list of TCP IPv4 network connections opened by pid.
func Connections(pid int32) ([]ConnectionStat, error) {
	root := "/proc"
	var err error
	var inodes map[string][]inodeMap
	inodes, err = getProcInodes(root, pid, 0)
	if len(inodes) == 0 {
		// no connection for the pid
		return []ConnectionStat{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cound not get pid(s), %d: %s", pid, err)
	}
	return statsFromInodes(root, pid, inodes)
}

func statsFromInodes(root string, pid int32, inodes map[string][]inodeMap) ([]ConnectionStat, error) {
	// check duplicate connections
	dupCheckMap := make(map[string]struct{})
	var ret []ConnectionStat

	var err error

	var path string
	var connKey string
	var ls []connTmp

	path = fmt.Sprintf("%s/%d/net/%s", root, pid, kindTCP4.filename)

	ls, err = processInet(path, kindTCP4, inodes, pid)
	if err != nil {
		return nil, err
	}
	for _, c := range ls {
		// Build TCP key to id the connection uniquely
		// socket type, src ip, src port, dst ip, dst port and state should be enough
		// to prevent duplications.
		connKey = fmt.Sprintf("%d-%s:%d-%s:%d-%s", c.sockType, c.laddr.IP, c.laddr.Port, c.raddr.IP, c.raddr.Port, c.status)
		if _, ok := dupCheckMap[connKey]; ok {
			continue
		}

		conn := ConnectionStat{
			Fd:     c.fd,
			Family: c.family,
			Type:   c.sockType,
			Laddr:  c.laddr,
			Raddr:  c.raddr,
			Status: c.status,
			Pid:    c.pid,
		}
		// fetch process owner Real, effective, saved set, and filesystem UIDs
		proc := process{Pid: conn.Pid}
		conn.Uids, _ = proc.getUids()

		ret = append(ret, conn)
		dupCheckMap[connKey] = struct{}{}
	}

	return ret, nil
}

// getProcInodes returnes fd of the pid.
func getProcInodes(root string, pid int32, max int) (map[string][]inodeMap, error) {
	ret := make(map[string][]inodeMap)

	dir := fmt.Sprintf("%s/%d/fd", root, pid)
	f, err := os.Open(dir)
	if err != nil {
		return ret, err
	}
	defer f.Close()
	files, err := f.Readdir(max)
	if err != nil {
		return ret, err
	}
	for _, fd := range files {
		inodePath := fmt.Sprintf("%s/%d/fd/%s", root, pid, fd.Name())

		inode, err := os.Readlink(inodePath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(inode, "socket:[") {
			continue
		}
		// the process is using a socket
		l := len(inode)
		inode = inode[8 : l-1]
		_, ok := ret[inode]
		if !ok {
			ret[inode] = make([]inodeMap, 0)
		}
		fd, err := strconv.Atoi(fd.Name())
		if err != nil {
			continue
		}

		i := inodeMap{
			pid: pid,
			fd:  uint32(fd),
		}
		ret[inode] = append(ret[inode], i)
	}
	return ret, nil
}

// Note: the following is based off process_linux structs and methods
// we need these to fetch the owner of a process ID
type process struct {
	Pid  int32 `json:"pid"`
	uids []int32
}

// getUids returns user ids of the process as a slice of the int
func (p *process) getUids() ([]int32, error) {
	err := p.fillFromStatus()
	if err != nil {
		return []int32{}, err
	}
	return p.uids, nil
}

// fillFromStatus Get status from /proc/(pid)/status
func (p *process) fillFromStatus() error {
	pid := p.Pid
	statPath := "/proc/" + strconv.Itoa(int(pid)) + "/status"
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		value := tabParts[1]
		switch strings.TrimRight(tabParts[0], ":") {
		case "Uid":
			p.uids = make([]int32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					return err
				}
				p.uids = append(p.uids, int32(v))
			}
		}
	}
	return nil
}

// decodeAddress decode address represents addr in proc/net/*
// ex:
// "0500000A:0016" -> "10.0.0.5", 22
func decodeAddress(family uint32, src string) (Addr, error) {
	t := strings.Split(src, ":")
	if len(t) != 2 {
		return Addr{}, fmt.Errorf("does not contain port, %s", src)
	}
	addr := t[0]
	port, err := strconv.ParseInt("0x"+t[1], 0, 64)
	if err != nil {
		return Addr{}, fmt.Errorf("invalid port, %s", src)
	}
	decoded, err := hex.DecodeString(addr)
	if err != nil {
		return Addr{}, fmt.Errorf("decode error, %s", err)
	}
	var ip net.IP
	// Assumes this is little_endian
	if family == syscall.AF_INET {
		ip = net.IP(Reverse(decoded))
	}
	return Addr{
		IP:   ip.String(),
		Port: uint32(port),
	}, nil
}

func processInet(file string, kind netConnectionKindType, inodes map[string][]inodeMap, filterPid int32) ([]connTmp, error) {

	if !utils.PathExists(file) {
		// if file path is not correct, return empty.
		return []connTmp{}, nil
	}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(contents, []byte("\n"))

	var ret []connTmp
	// skip first line
	for _, line := range lines[1:] {
		l := strings.Fields(string(line))
		if len(l) < 10 {
			continue
		}
		laddr := l[1]
		raddr := l[2]
		status := l[3]
		inode := l[9]
		pid := int32(0)
		fd := uint32(0)
		i, exists := inodes[inode]
		if exists {
			pid = i[0].pid
			fd = i[0].fd
		}
		if filterPid > 0 && filterPid != pid {
			continue
		}
		if kind.sockType == syscall.SOCK_STREAM {
			status = TCPStatuses[status]
		} else {
			status = "NONE"
		}
		la, err := decodeAddress(kind.family, laddr)
		if err != nil {
			continue
		}
		ra, err := decodeAddress(kind.family, raddr)
		if err != nil {
			continue
		}

		ret = append(ret, connTmp{
			fd:       fd,
			family:   kind.family,
			sockType: kind.sockType,
			laddr:    la,
			raddr:    ra,
			status:   status,
			pid:      pid,
		})
	}

	return ret, nil
}

// Reverse reverses array of bytes.
func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
