// +build linux

package cpu

import (
	"errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"
)

func getTotalCPU() (uint64, error) {
	cpuStats, err := ioutil.ReadFile("/proc/stat")
	cpuTotalStat := strings.Split(strings.Split(string(cpuStats), "\n")[0], " ")[1:]
	var timeTotal uint64
	for _, cpuTime := range cpuTotalStat {
		timeField, _ := strconv.ParseUint(cpuTime, 10, 64)
		timeTotal = timeTotal + timeField
	}
	return timeTotal, err
}

func getProcStat(pid int) (utime, stime uint64, err error) {
	procStatFileBytes, err := ioutil.ReadFile(path.Join("/proc", strconv.Itoa(pid), "stat"))
	info := strings.Split(string(procStatFileBytes), " ")

	if len(info) <= 1 || err != nil {
		return uint64(0), uint64(0), errors.New("Can't find process with this PID: " + strconv.Itoa(pid))
	}
	utime, _ = strconv.ParseUint(info[14], 10, 64)
	stime, _ = strconv.ParseUint(info[15], 10, 64)

	return utime, stime, err
}

func countCPUs() float64 {
	cpuInfoFileBytes, _ := ioutil.ReadFile("/proc/cpuinfo")
	cpuInfoFileLines := strings.Split(string(cpuInfoFileBytes), "\n")
	var sum float64
	for _, line := range cpuInfoFileLines {
		if !strings.HasPrefix(line, "processor") {
			continue
		}
		sum = sum + 1
	}
	return sum
}

func GetProcessCPUPerformance(pid int) (float64, error) {
	var utimeBefore, stimeBefore, utimeAfter, stimeAfter uint64
	// get total time(of the system), utime and stime (of the process stat file) before time interval.
	timeTotalBefore, err := getTotalCPU()
	utimeBefore, stimeBefore, err = getProcStat(pid)
	// sleep for interval of 2 seconds for the proc file system to refresh.
	time.Sleep(time.Second * 2)
	// get total time(of the system), utime and stime (of the process stat file) after time interval.
	timeTotalAfter, err := getTotalCPU()
	utimeAfter, stimeAfter, err = getProcStat(pid)

	ret := countCPUs() * float64((utimeAfter+stimeAfter)-(utimeBefore+stimeBefore)) * 100 / float64(timeTotalAfter-timeTotalBefore)

	return ret, err
}
