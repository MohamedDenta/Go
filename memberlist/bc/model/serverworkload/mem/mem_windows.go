// +build windows

package mem

import (

	// "syscall"
	// "unsafe"

	"fmt"
	"strconv"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// var (
// 	procGetProcessMemoryInfo = utils.ModPsapi.NewProc("GetProcessMemoryInfo")
// )

// type PROCESS_MEMORY_COUNTERS struct {
// 	CB                         uint32
// 	PageFaultCount             uint32
// 	PeakWorkingSetSize         uintptr
// 	WorkingSetSize             uintptr
// 	QuotaPeakPagedPoolUsage    uintptr
// 	QuotaPagedPoolUsage        uintptr
// 	QuotaPeakNonPagedPoolUsage uintptr
// 	QuotaNonPagedPoolUsage     uintptr
// 	PagefileUsage              uintptr
// 	PeakPagefileUsage          uintptr
// }

// type Win32_Processor struct {
// 	Name        string
// 	ProcessorID *string
// }

// // GetProcessMemoryPerformance
// func GetProcessMemoryUtilization(pid int) (uint64, error) {
// 	var pmc PROCESS_MEMORY_COUNTERS
// 	hProcess, _ := syscall.OpenProcess(uint32(3), false, uint32(pid))
// 	// err := GetProcessMemoryInfo(hProcess, &pmc, unsafe.Sizeof(&pmc))
// 	r1, _, e1 := syscall.Syscall(procGetProcessMemoryInfo.Addr(), 3, uintptr(hProcess), uintptr(unsafe.Pointer(&pmc)), uintptr(unsafe.Sizeof(&pmc)))
// 	var err error
// 	if r1 == 0 {
// 		if e1 != 0 {
// 			err = error(e1)
// 		} else {
// 			err = syscall.EINVAL
// 		}
// 	}
// 	fmt.Println("mem err??", err)
// 	fmt.Println("Working Set Size:", uint64(pmc.PagefileUsage))

// 	return uint64(pmc.PagefileUsage), err
// }

func GetProcessMemoryUtilization(pid int) (uint64, error) {
	// return InfoWithContext(context.Background(), uint32(pid))
	// init COM, oh yeah
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	defer unknown.Release()

	wmi, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, _ := oleutil.CallMethod(wmi, "ConnectServer")
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := fmt.Sprintf("SELECT WorkingSetPrivate FROM Win32_PerfRawData_PerfProc_Process Where IDProcess=%v", pid)
	// result is a SWBemObjectSet
	resultRaw, _ := oleutil.CallMethod(service, "ExecQuery", query)
	result := resultRaw.ToIDispatch()
	defer result.Release()

	countVar, _ := oleutil.GetProperty(result, "Count")
	count := int(countVar.Val)

	for i := 0; i < count; i++ {
		// item is a SWbemObject, but really a Win32_Process
		itemRaw, _ := oleutil.CallMethod(result, "ItemIndex", i)
		item := itemRaw.ToIDispatch()
		defer item.Release()

		mem, err := oleutil.GetProperty(item, "WorkingSetPrivate")
		fmt.Println("err for ole?", err)
		// fmt.Printf("mem: %v %T\n", mem.Value(), mem.Value())
		// procPrivateMem, _ := strconv.ParseUint(mem.Value(), 10, 64)
		// fmt.Println("memory:", (procPrivateMem * 4096))
		if memRetInt, ok := mem.Value().(int32); ok {
			// fmt.Printf("ret mem: %v %T", ret, ret)
			return uint64(memRetInt), err
		} else if memRet, ok := mem.Value().(string); ok {
			memRet, _ := strconv.ParseUint(memRet, 10, 64)
			return memRet, err
		}
		// asString, _ := oleutil.GetProperty(item, "IDProcess")
		// var parsedString int
		// // fmt.Println("err for ole?", err)
		// // fmt.Printf("pid type: %v %T\n", asString.Value(), asString.Value())
		// if asStringCpy, ok := asString.Value().(int32); ok {
		// 	parsedString = int(asStringCpy) // strconv.Atoi(asString)
		// }

		// if parsedString == pid {
		// }
	}
	return uint64(0), err
}

// func InfoWithContext(ctx context.Context, pid uint32) (uint64, error) {
// 	var ret uint64
// 	var dst []Win32_Processor
// 	q := wmi.CreateQuery(&dst, "")
// 	if err := WMIQueryWithContext(ctx, q, &dst); err != nil {
// 		return ret, err
// 	}

// 	// var procID string
// 	for _, l := range dst {
// 		if l.ProcessorID != nil {
// 			fmt.Println("PerfProc pid:", *l.ProcessorID)
// 		}

// 	}

// 	return ret, nil
// }

// // WMIQueryWithContext - wraps wmi.Query with a timed-out context to avoid hanging
// func WMIQueryWithContext(ctx context.Context, query string, dst interface{}, connectServerArgs ...interface{}) error {
// 	timeout := 3 * time.Second
// 	if _, ok := ctx.Deadline(); !ok {
// 		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
// 		defer cancel()
// 		ctx = ctxTimeout
// 	}

// 	errChan := make(chan error, 1)
// 	go func() {
// 		errChan <- wmi.Query(query, dst, connectServerArgs...)
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		return ctx.Err()
// 	case err := <-errChan:
// 		return err
// 	}
// }

// // GetProcessMemoryInfo
// func GetProcessMemoryInfo(handle syscall.Handle, memCounters *PROCESS_MEMORY_COUNTERS, cb uintptr) (err error) {

// 	return
// }
