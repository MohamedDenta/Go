package serverworkload

import (
	"fmt"
	"../serverworkload/cpu"
	"../serverworkload/disk"
	"../serverworkload/mem"
	"../serverworkload/net"
	"../serverworkload/utils"
	"net/http"
	"os"
	"strconv"
	"syscall"
)

var serverPID = syscall.Getpid() // make toml configuration or set it manually like this int(2865)
func GetCPUWorkloadPrecentage() string {
	cpuServerUsage, _ := cpu.GetProcessCPUPerformance(serverPID)
	cpuServerUsagestr := strconv.FormatFloat(cpuServerUsage, 'f', 2, 64) + "%"
	return cpuServerUsagestr
}

// AllStats controller for all server resources stats
func AllStats(w http.ResponseWriter, r *http.Request) {

	
	cpuServerUsage, _ := cpu.GetProcessCPUPerformance(serverPID)
	fmt.Println("cpu usage:", cpuServerUsage)

	memUsage, _ := mem.GetProcessMemoryUtilization(serverPID)

	hostname, _ := os.Hostname()

	networkStats, _ := net.IOCounters()
	var networkStatsOutput []map[string]interface{}
	for _, counter := range networkStats {
		networkStatsOutput = append(networkStatsOutput, map[string]interface{}{
			"name":        counter.Name,
			"bytesSent":   utils.FormatBytes(counter.BytesSent, 2),
			"bytesRecv":   utils.FormatBytes(counter.BytesRecv, 2),
			"packetsSent": utils.FormatBytes(counter.PacketsSent, 2),
			"packetsRecv": utils.FormatBytes(counter.PacketsRecv, 2),
		})
	}

	// TODO: adjust the connections with node port from the toml file
	connections, _ := net.Connections(int32(serverPID))

	networkData := map[string]interface{}{
		"hostname":                       hostname,
		"nodeestablishedrequestscount": len(connections),
		"networkinterfacesstats":        networkStatsOutput,
	}

	partitions, _ := disk.Partitions()
	var storageStatsOutput []map[string]interface{}
	// loop over partitions and get stats for every partition
	for _, part := range partitions {
		u, _ := disk.Usage(part.Mountpoint)
		storageStatsOutput = append(storageStatsOutput, map[string]interface{}{
			"path":        u.Path,
			"fstype":      u.Fstype,
			"total":       utils.FormatBytes(u.Total, 2),
			"free":        utils.FormatBytes(u.Free, 2),
			"used":        utils.FormatBytes(u.Used, 2),
			"usedPercent": fmt.Sprintf("%.2f%%", u.UsedPercent),
		})
	}

	utils.Respond(w, http.StatusOK, map[string]interface{}{
		"nodecpuusage":             fmt.Sprintf("%.2f%%", cpuServerUsage),
		"nodememoryusage":          utils.FormatBytes(memUsage, 2),
		"networkstats":              networkData,
		"storagestats":              storageStatsOutput,
		"currenttransactionscount": 20, // new_transaction.CountTransaction(),
	})
}

// CPUStats controller for the CPU Usage route
func CPUStats(w http.ResponseWriter, r *http.Request) {
	cpuServerUsage, _ := cpu.GetProcessCPUPerformance(serverPID)

	utils.Respond(w, http.StatusOK, map[string]interface{}{
		"nodecpuusage": fmt.Sprintf("%.2f%%", cpuServerUsage),
	})
}

// MemoryStats controller for the Memory statistics and usage
func MemoryStats(w http.ResponseWriter, r *http.Request) {
	memUsage, _ := mem.GetProcessMemoryUtilization(serverPID)

	// return the response for the client
	utils.Respond(w, http.StatusOK, map[string]interface{}{
		"nodememoryusage": utils.FormatBytes(memUsage, 2),
	})
}

// NetworkStats controller for the network statistics and usage
func NetworkStats(w http.ResponseWriter, r *http.Request) {
	networkStats, _ := net.IOCounters()
	var networkStatsOutput []map[string]interface{}
	for _, counter := range networkStats {
		networkStatsOutput = append(networkStatsOutput, map[string]interface{}{
			"name":        counter.Name,
			"bytesSent":   utils.FormatBytes(counter.BytesSent, 2),
			"bytesRecv":   utils.FormatBytes(counter.BytesRecv, 2),
			"packetsSent": utils.FormatBytes(counter.PacketsSent, 2),
			"packetsRecv": utils.FormatBytes(counter.PacketsRecv, 2),
		})
	}

	hostname, _ := os.Hostname()

	connections, _ := net.Connections(int32(serverPID))
	//fmt.Println("connections:", connections)

	// return the response for the client
	utils.Respond(w, http.StatusOK, map[string]interface{}{
		"hostname":                       hostname,
		"nodeestablishedrequestscount": len(connections),
		"networkstats":                   networkStatsOutput,
	})
}

// StorageStats controller for the Storage statistics and usage
func StorageStats(w http.ResponseWriter, r *http.Request) {
	partitions, _ := disk.Partitions()

	var storageStatsOutput []map[string]interface{}
	// loop over partitions and get stats for every partition
	for _, part := range partitions {
		u, _ := disk.Usage(part.Mountpoint)
		storageStatsOutput = append(storageStatsOutput, map[string]interface{}{
			"path":        u.Path,
			"fstype":      u.Fstype,
			"total":       utils.FormatBytes(u.Total, 2),
			"free":        utils.FormatBytes(u.Free, 2),
			"used":        utils.FormatBytes(u.Used, 2),
			"usedPercent": fmt.Sprintf("%.2f%%", u.UsedPercent),
		})
	}

	// return the response for the client
	utils.Respond(w, http.StatusOK, map[string]interface{}{
		"partitionsstats": storageStatsOutput,
	})
}

// CurrentTransactionsCount controller for the number of current transactions in transaction pool
func CurrentTransactionsCount(w http.ResponseWriter, r *http.Request) {

	utils.Respond(w, http.StatusOK, map[string]interface{}{
		"currenttransactionscount": 20, //new_transaction.CountTransaction(),
	})
}
