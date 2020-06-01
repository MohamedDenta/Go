// +build windows

package cpu

import (
	"fmt"
	"runtime"
	"strconv"
	"unsafe"

	"../utils"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

var (
	procGetSystemTimes   = utils.Modkernel32.NewProc("GetSystemTimes")
	lastCPUSystemPercent LastPercent
	serverPID            = int(4056)
	cpusNumber           = runtime.NumCPU()
)

func init() {
	lastCPUSystemPercent.Lock()
	lastCPUSystemPercent.LastCPUTimes, _ = systemTimes()
	lastCPUSystemPercent.Unlock()
}

// getSystemCPUPerformance
func getSystemCPUPerformance() (float64, error) {
	percent, err := systemPercentUsedFromLastCall()
	return percent[0], err
}

// GetProcessCPUPerformance
func GetProcessCPUPerformance(pid int) (float64, error) {
	cpuTimes := processTimes(pid)

	return (cpuTimes.PercentProcessTime / float64(cpusNumber)), nil
}

// ProcessTimes
func processTimes(pid int) ProcessPercent {
	// init COM
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer unknown.Release()

	wmi, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, _ := oleutil.CallMethod(wmi, "ConnectServer")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := fmt.Sprintf("SELECT * FROM Win32_PerfFormattedData_PerfProc_Process Where IDProcess=%v", pid)
	// result is a SWBemObjectSet
	resultRaw, _ := oleutil.CallMethod(service, "ExecQuery", query)
	// fmt.Println("query err?", err)
	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, _ := oleutil.GetProperty(result, "Count")
	count := int(countVar.Val)
	// fmt.Println("process count:", count)

	ret := ProcessPercent{}
	for i := 0; i < count; i++ {
		// item is a SWbemObject, but really a Win32_Process
		itemRaw, _ := oleutil.CallMethod(result, "ItemIndex", i)
		item := itemRaw.ToIDispatch()
		defer item.Release()

		percentProcessorTime, _ := oleutil.GetProperty(item, "PercentProcessorTime")
		// fmt.Println("err for ole?", err)

		if percentProcessorInt, ok := percentProcessorTime.Value().(int32); ok {
			ret.PercentProcessTime = float64(percentProcessorInt)
			fmt.Println("PercentProcessorTime int:", float64(percentProcessorInt))
		} else if percentProcessorStr, ok := percentProcessorTime.Value().(string); ok {
			ret.PercentProcessTime, _ = strconv.ParseFloat(percentProcessorStr, 64)
			//fmt.Println("PercentProcessorTime: string", percentProcessorStr)
		}
	}
	return ret
}

// systemTimes
func systemTimes() ([]TimesStat, error) {
	var ret []TimesStat
	var lpIdleTime windows.Filetime
	var lpKernelTime windows.Filetime
	var lpUserTime windows.Filetime
	r, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&lpIdleTime)),
		uintptr(unsafe.Pointer(&lpKernelTime)),
		uintptr(unsafe.Pointer(&lpUserTime)))
	if r == 0 {
		return ret, windows.GetLastError()
	}

	LOT := float64(0.0000001)
	HIT := (LOT * 4294967296.0)
	idle := ((HIT * float64(lpIdleTime.HighDateTime)) + (LOT * float64(lpIdleTime.LowDateTime)))
	user := ((HIT * float64(lpUserTime.HighDateTime)) + (LOT * float64(lpUserTime.LowDateTime)))
	kernel := ((HIT * float64(lpKernelTime.HighDateTime)) + (LOT * float64(lpKernelTime.LowDateTime)))
	system := (kernel - idle)

	ret = append(ret, TimesStat{
		CPU:    "cpu-total",
		Idle:   float64(idle),
		User:   float64(user),
		System: float64(system),
	})
	return ret, nil
}

// systemPercentUsedFromLastCall
func systemPercentUsedFromLastCall() ([]float64, error) {
	cpuTimes, err := systemTimes()
	if err != nil {
		return nil, err
	}
	var lastTimes []TimesStat
	lastCPUSystemPercent.Lock()
	defer lastCPUSystemPercent.Unlock()

	lastTimes = lastCPUSystemPercent.LastCPUTimes
	lastCPUSystemPercent.LastCPUTimes = cpuTimes
	if lastTimes == nil {
		return nil, fmt.Errorf("error getting times for cpu percent. lastTimes was nil")
	}

	return calculateAllBusy(lastTimes, cpuTimes)
}

// getAllBusy
func getAllBusy(t TimesStat) (float64, float64) {
	busy := t.User + t.System + t.Nice + t.Iowait + t.Irq +
		t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
	return busy + t.Idle, busy
}

// calculateBusy
func calculateBusy(t1, t2 TimesStat) float64 {
	t1All, t1Busy := getAllBusy(t1)
	t2All, t2Busy := getAllBusy(t2)

	if t2Busy <= t1Busy {
		return 0
	}
	if t2All <= t1All {
		return 1
	}
	return (t2Busy - t1Busy) / (t2All - t1All) * 100
}

// calculateAllBusy
func calculateAllBusy(t1, t2 []TimesStat) ([]float64, error) {
	// Make sure the CPU measurements have the same length.
	if len(t1) != len(t2) {
		return nil, fmt.Errorf(
			"received two CPU counts: %d != %d",
			len(t1), len(t2),
		)
	}

	ret := make([]float64, len(t1))
	for i, t := range t2 {
		ret[i] = calculateBusy(t1[i], t)
	}
	return ret, nil
}
