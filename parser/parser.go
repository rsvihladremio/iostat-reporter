package parser

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"time"
)

// ParseIostatOutput parses the full iostat -x -c output into structured data.
func ParseIostatOutput(data []byte) (ParsedData, error) {
	parsed := ParsedData{
		Devices: make(map[string][]DeviceStats),
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// parse timestamp exactly MM/DD/YY HH:MM:SS
		ts, err := time.Parse("01/02/06 15:04:05", line)
		if err != nil {
			// not a timestamp line
			continue
		}

		// find and consume avg-cpu header
		var cpuHeader []string
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(text, "avg-cpu:") {
				parts := strings.Fields(text)
				if len(parts) > 1 {
					cpuHeader = parts[1:]
				}
				break
			}
		}
		// find CPU values
		var cpuValues []string
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}
			cpuValues = strings.Fields(text)
			break
		}
		// map CPU by header
		cpuMap := make(map[string]float64, len(cpuHeader))
		for i, h := range cpuHeader {
			key := strings.TrimPrefix(h, "%")
			if i < len(cpuValues) {
				v, err := strconv.ParseFloat(cpuValues[i], 64)
				if err != nil {
					return parsed, err
				}
				cpuMap[key] = v
			}
		}
		parsed.CPUs = append(parsed.CPUs, CPUStats{
			Timestamp: ts,
			User:      cpuMap["user"],
			Nice:      cpuMap["nice"],
			System:    cpuMap["system"],
			Iowait:    cpuMap["iowait"],
			Steal:     cpuMap["steal"],
			Idle:      cpuMap["idle"],
		})

		// find device header
		var devHeader []string
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}
			if strings.HasPrefix(text, "Device") {
				devHeader = strings.Fields(text)
				break
			}
		}
		// parse device lines
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				break
			}
			fields := strings.Fields(text)
			devMap := make(map[string]float64, len(devHeader))
			for i, h := range devHeader {
				if i == 0 || i >= len(fields) {
					continue
				}
				key := strings.TrimPrefix(h, "%")
				v, err := strconv.ParseFloat(fields[i], 64)
				if err != nil {
					return parsed, err
				}
				devMap[key] = v
			}
			dev := DeviceStats{
				Timestamp:         ts,
				Name:              fields[0],
				ReadsPerSec:       devMap["r/s"],
				ReadKBPerSec:      devMap["rkB/s"],
				ReadMergedPerSec:  devMap["rrqm/s"],
				ReadPctMerged:     devMap["rrqm"],
				ReadAwaitMs:       devMap["r_await"],
				ReadReqSzKB:       devMap["rareq-sz"],
				WritesPerSec:      devMap["w/s"],
				WriteKBPerSec:     devMap["wkB/s"],
				WriteMergedPerSec: devMap["wrqm/s"],
				WritePctMerged:    devMap["wrqm"],
				WriteAwaitMs:      devMap["w_await"],
				WriteReqSzKB:      devMap["wareq-sz"],
				// aqu-sz → queue length
				QueueSize: devMap["aqu-sz"],
			}
			parsed.Devices[dev.Name] = append(parsed.Devices[dev.Name], dev)
		}
	}

	if err := scanner.Err(); err != nil {
		return parsed, err
	}
	return parsed, nil
}

// parseTimestamp parses lines like "09/04/24 12:07:20".
func parseTimestamp(line string) (time.Time, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return time.Time{}, errors.New("not a timestamp")
	}
	stamp := parts[0] + " " + parts[1]
	// format: MM/DD/YY HH:MM:SS
	return time.Parse("01/02/06 15:04:05", stamp)
}

// parseCPUStats takes a timestamp and six string fields, returns a CPUStats.
func parseCPUStats(ts time.Time, fields []string) (CPUStats, error) {
	if len(fields) != 6 {
		return CPUStats{}, errors.New("cpu stats: wrong number of fields")
	}
	vals := make([]float64, 6)
	for i := 0; i < 6; i++ {
		v, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return CPUStats{}, err
		}
		vals[i] = v
	}
	return CPUStats{
		Timestamp: ts,
		User:      vals[0],
		Nice:      vals[1],
		System:    vals[2],
		Iowait:    vals[3],
		Steal:     vals[4],
		Idle:      vals[5],
	}, nil
}

// parseDeviceStats takes a timestamp and at least 13 fields, returns a DeviceStats.
func parseDeviceStats(ts time.Time, fields []string) (DeviceStats, error) {
	if len(fields) < 13 {
		return DeviceStats{}, errors.New("device stats: wrong number of fields")
	}
	// helper to parse one field
	get := func(idx int) (float64, error) {
		return strconv.ParseFloat(fields[idx], 64)
	}

	rps, err := get(1)
	if err != nil {
		return DeviceStats{}, err
	}
	rkb, err := get(2)
	if err != nil {
		return DeviceStats{}, err
	}
	rrqm, err := get(3)
	if err != nil {
		return DeviceStats{}, err
	}
	rpct, err := get(4)
	if err != nil {
		return DeviceStats{}, err
	}
	raw, err := get(5)
	if err != nil {
		return DeviceStats{}, err
	}
	rsz, err := get(6)
	if err != nil {
		return DeviceStats{}, err
	}
	wps, err := get(7)
	if err != nil {
		return DeviceStats{}, err
	}
	wkb, err := get(8)
	if err != nil {
		return DeviceStats{}, err
	}
	wrqm, err := get(9)
	if err != nil {
		return DeviceStats{}, err
	}
	wpct, err := get(10)
	if err != nil {
		return DeviceStats{}, err
	}
	waw, err := get(11)
	if err != nil {
		return DeviceStats{}, err
	}
	wsz, err := get(12)
	if err != nil {
		return DeviceStats{}, err
	}
	return DeviceStats{
		Timestamp:         ts,
		Name:              fields[0],
		ReadsPerSec:       rps,
		ReadKBPerSec:      rkb,
		ReadMergedPerSec:  rrqm,
		ReadPctMerged:     rpct,
		ReadAwaitMs:       raw,
		ReadReqSzKB:       rsz,
		WritesPerSec:      wps,
		WriteKBPerSec:     wkb,
		WriteMergedPerSec: wrqm,
		WritePctMerged:    wpct,
		WriteAwaitMs:      waw,
		WriteReqSzKB:      wsz,
	}, nil
}
