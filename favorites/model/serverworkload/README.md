## the serverworkload module:
It's a web service that will be called to check the work load and resources Stats on the node server.
    <br> The web service will get the following data and evaluate server performance:
* Amount of load (in the form of client requests).
* Processing and communications resources (RAM, Network Stats).
* Network info like the server IP.
* Consumed storage resources.
* Number of current transactions in the transaction pool.

### The functions of the serverworkload:
* <b> GetProcessCPUPerformance</b> ( pid <b>int</b> ):<br>
    will return the process CPU usage percentage as <b>float64</b> by using the pid to look into the /proc file system
    in linux and calculate the CPU usage.
* <b> GetProcessMemoryUtilization</b> ( pid <b>int</b> ):<br>
    will return the process private memory usage as number of bytes <b>uint64</b> by using the pid to look into the /proc file system
    in linux and calculate the private memory.
* <b> IOCounters</b> ():<br>
    will return the network I/O statistics (<b>IOCountersStat</b> struct) for wlan and ethernet
     network interface installed on the system.
* <b> Connections</b> ( pid <b>int32</b> ):<br>
    Connections Return a list of TCP IPv4 network connections opened by a process.
* <b> Partitions</b> ():<br>
    will returns disk partitions (slice of <b>PartitionStat</b> struct). returns physical devices only
    (e.g. hard disks, cd-rom drives, USB keys).

### The web serivce API:
* <b>CPUStats</b> endpoint `/stats/cpu` :<br>
    A controller to get the node server CPU utilization percentage.<br>
    * Allowed methods: `GET`, `OPTIONS` <br>
    * Response data: <br>
        <pre>{
        "node cpu usage":                      string,
      }</pre>
    * Example response: <br>
        <pre>{
         "node cpu usage":                     "8.05%",
      }</pre>
* <b>MemoryStats</b> endpoint `/stats/memory` :<br>
    A controller to get the node server memory utilization with formatted byte number.<br>
    * Allowed methods: `GET`, `OPTIONS` <br>
    * Response data: <br>
        <pre>{
         "node memory usage":                     string,
      }</pre>
    * Example response: <br>
        <pre>{
         "node memory usage":                     "297.03 MB",
      }</pre>
* <b>NetworkStats</b> endpoint `/stats/network` :<br>
    A controller to get the TCP IPv4 network statistics in formatted byte number.<br>
    * Allowed methods: `GET`, `OPTIONS` <br>
    * Response data: <br>
        <pre>{
         "host name":                          string,
         "node established requests count":    int,
         "network interfaces stats":           [interface object],
      }</pre>
      *  the "network interfaces stats" are a list of interface stats objects, and each object are:<br>
            <pre>{
             "Name":                               string,
             "system bytes sent":                  string,
             "system bytes received":              string,
             "system packets sent":                string,
             "system packets received":            string
         }</pre>
     * Example response: <br>
         <pre> {
          "host name": "dhcppc6",
          "network interfaces stats": [
            {
              "bytesRecv": "0 Bytes",
              "bytesSent": "0 Bytes",
              "name": "virbr0-nic",
              "packetsRecv": "0 Bytes",
              "packetsSent": "0 Bytes"
            },
            {
              "bytesRecv": "6.937 GB",
              "bytesSent": "282.84 MB",
              "name": "wlp2s0b1",
              "packetsRecv": "4.97 MB",
              "packetsSent": "2.83 MB"
            },
            {
              "bytesRecv": "0 Bytes",
              "bytesSent": "0 Bytes",
              "name": "virbr0",
              "packetsRecv": "0 Bytes",
              "packetsSent": "0 Bytes"
            },
            {
              "bytesRecv": "0 Bytes",
              "bytesSent": "0 Bytes",
              "name": "eno1",
              "packetsRecv": "0 Bytes",
              "packetsSent": "0 Bytes"
            }
          ],
          "node established requests count": 10
        }</pre>
* <b>StorageStats</b> endpoint `/stats/storage` :<br>
    A controller to get the Storage statistics and utilization for each partition with formatted byte number.<br>
    * Allowed methods: `GET`, `OPTIONS` <br>
    * Response data: <br>
        <pre>{
         "partitions stats":                   [partition object]
      }</pre>
      *  the "partitions stats" are a list of partition objects, and each object are:<br>
         <pre>{
             "path":                               string,
             "fstype":                             string,
             "total":                              string,
             "free":                               string,
             "used":                               string,
             "usedPercent":                        string,
         }</pre>
     * Example response: <br>
         <pre>{
         "storage stats": [
             {
               "free": "24.683 GB",
               "fstype": "ext2/ext3",
               "path": "/",
               "total": "52.710 GB",
               "used": "25.326 GB",
               "usedPercent": "50.64%"
             },
             {
               "free": "14.382 GB",
               "fstype": "ext2/ext3",
               "path": "/home",
               "total": "73.848 GB",
               "used": "55.692 GB",
               "usedPercent": "79.48%"
             },
           ]
       }</pre>
* <b>CurrentTransactionsCount</b> endpoint `/stats/current-transactions-count` :<br>
    A controller for all server resources stats.<br>
    * Allowed methods: `GET`, `OPTIONS` <br>
    * Response data: <br>
        <pre>{
         "current transactions count":         int,
      }</pre>
* <b>AllStats</b> endpoint `/stats/` :<br>
    A controller for all server resources stats.<br>
    * Allowed methods: `GET`, `OPTIONS` <br>
    * Response data: <br>
        <pre>{
          "node cpu usage":                     string,
          "node memory usage":                  string,
          "network stats":                      network object,
          "storage stats":                      [partition object],
          "current transactions count":         int,
      }</pre>

### Requirements:
* go version: 1.11.4.
* `go get "github.com/go-ole/go-ole"`
* `go get "github.com/go-ole/go-ole/oleutil"`
* `go get "golang.org/x/sys/windows"`
* `go get "golang.org/x/sys/unix"`

### Important Notes:

* you can use this module's functions without the web service api, or use its web service api. <br> 
  either way you will have to provide the ID of the process you want to monitor. you can use <br>
  `syscall.Getpid()` to get the current process ID.
* on windows there's a bug in the `GetProcessCPUPerformance` function, it's a logical bug i think.