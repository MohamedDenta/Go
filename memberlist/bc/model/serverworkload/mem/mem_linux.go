// +build linux

package mem

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetProcessMemoryPerformance
func GetProcessMemoryUtilization(pid int) (uint64, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := bufio.NewScanner(f)
	r.Scan() // read first line
	line := strings.Split(r.Text(), " ")
	if len(line) <= 1 {
		return 0, errors.New("Can't find RSS and Shared memory byte counters for this PID: " + string(pid))
	}
	rss, _ := strconv.ParseUint(line[1], 10, 64)
	shm, _ := strconv.ParseUint(line[2], 10, 64)
	pageSize := uint64(4096)
	//pageSizeStdout, err := exec.Command("getconf", "PAGESIZE").Output()
	//if err == nil {
	//	pageSize, err = strconv.ParseUint(string(pageSizeStdout), 10, 64)
	//}
	return (rss * pageSize) - (shm * pageSize), nil
}
