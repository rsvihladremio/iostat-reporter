package parser

import "time"

// CPUStats holds the average‐CPU metrics for one timestamp.
type CPUStats struct {
	Timestamp time.Time
	User      float64
	Nice      float64
	System    float64
	Iowait    float64
	Steal     float64
	Idle      float64
}

// DeviceStats holds the per‐device I/O metrics for one timestamp.
type DeviceStats struct {
	Timestamp        time.Time
	Name             string
	ReadsPerSec      float64
	ReadKBPerSec     float64
	ReadMergedPerSec float64
	ReadPctMerged    float64
	ReadAwaitMs      float64
	ReadReqSzKB      float64

	WritesPerSec      float64
	WriteKBPerSec     float64
	WriteMergedPerSec float64
	WritePctMerged    float64
	WriteAwaitMs      float64
	WriteReqSzKB      float64
}

// ParsedData is the top‐level result of parsing iostat output.
type ParsedData struct {
	CPUs    []CPUStats
	Devices map[string][]DeviceStats
}
