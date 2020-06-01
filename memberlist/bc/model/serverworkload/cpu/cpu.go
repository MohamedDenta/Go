package cpu

import (
	"sync"
)

// TimesStat contains the amounts of time the CPU has spent performing different
// kinds of work. Time units are in USER_HZ or Jiffies (typically hundredths of
// a second). It is based on linux /proc/stat file.
type TimesStat struct {
	CPU       string  `json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guestNice"`
	Stolen    float64 `json:"stolen"`
	Now       float64 `json:"systemNow"`
}

type LastPercent struct {
	sync.Mutex
	LastCPUTimes []TimesStat
}

type ProcessPercent struct {
	PercentProcessTime float64
	TimeStamp          float64
}

type lastProcessPercent struct {
	sync.Mutex
	ProcessPercentUtil ProcessPercent
}
