package reporter

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rsvihladremio/iostat-reporter/parser"
)

// makeDummyParsedData returns a minimal ParsedData for testing.
func makeDummyParsedData() parser.ParsedData {
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	return parser.ParsedData{
		CPUs: []parser.CPUStats{
			{Timestamp: now, User: 10, System: 5, Idle: 85, Iowait: 0, Nice: 0, Steal: 0},
			{Timestamp: now.Add(time.Second), User: 20, System: 2, Idle: 78, Iowait: 0, Nice: 0, Steal: 0},
		},
		Devices: map[string][]parser.DeviceStats{
			"sda": {
				{Timestamp: now, ReadsPerSec: 1, WritesPerSec: 2, ReadKBPerSec: 1024, WriteKBPerSec: 2048, ReadAwaitMs: 0.5, WriteAwaitMs: 1.5, QueueSize: 0.1},
				{Timestamp: now.Add(time.Second), ReadsPerSec: 3, WritesPerSec: 4, ReadKBPerSec: 512, WriteKBPerSec: 1024, ReadAwaitMs: 0.2, WriteAwaitMs: 0.8, QueueSize: 0.05},
			},
		},
	}
}

func TestGenerateReport_Minimal(t *testing.T) {
	parsed := makeDummyParsedData()
	dir := t.TempDir()
	out := filepath.Join(dir, "report.html")

	if err := GenerateReport(parsed, out, "My Title", "Some metadata", "myfile.log", "abcdef1234567890", "v1.2.3"); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	html := string(data)

	// check title
	if !strings.Contains(html, "<h1") || !strings.Contains(html, "My Title") {
		t.Error("expected title My Title in output")
	}

	// check abbreviated hash
	if !strings.Contains(html, "abcdef..") {
		t.Error("expected abbreviated hash in output")
	}

	// extract CPU JSON and validate
	parts := strings.Split(html, "cpuChart.setOption(")
	if len(parts) < 2 {
		t.Fatal("cannot locate CPU JSON")
	}
	raw := strings.SplitN(parts[1], ");", 2)[0]
	jsonTxt := strings.TrimSuffix(raw, "\n")
	var cpuOpt map[string]interface{}
	if err := json.Unmarshal([]byte(jsonTxt), &cpuOpt); err != nil {
		t.Errorf("CPU JSON invalid: %v", err)
	}
	legend, ok := cpuOpt["legend"].(map[string]interface{})
	if !ok {
		t.Errorf("cpuOption.legend not a map")
	}
	dataArr, ok := legend["data"].([]interface{})
	if !ok || len(dataArr) != 6 {
		t.Errorf("unexpected legend data: %#v", legend["data"])
	}

	// check device name
	if !strings.Contains(html, "sda") {
		t.Error("device name sda not in output")
	}
}

// TestGenerateReport_EmptyData should generate a report with no CPUs or devices.
func TestGenerateReport_EmptyData(t *testing.T) {
	parsed := parser.ParsedData{}
	dir := t.TempDir()
	out := filepath.Join(dir, "empty.html")

	if err := GenerateReport(parsed, out, "Empty", "", "empty.log", "123456", "v0.0"); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	html := string(data)

	// Title is present
	if !strings.Contains(html, "Empty") {
		t.Error("expected title Empty in output")
	}

	// CPU chart data arrays should be empty
	parts := strings.Split(html, "cpuChart.setOption(")
	if len(parts) < 2 {
		t.Fatal("cannot locate CPU JSON")
	}
	raw := strings.SplitN(parts[1], ");", 2)[0]
	jsonTxt := strings.TrimSuffix(raw, "\n")
	var cpuOpt map[string]interface{}
	if err := json.Unmarshal([]byte(jsonTxt), &cpuOpt); err != nil {
		t.Errorf("CPU JSON invalid: %v", err)
	}
	series, ok := cpuOpt["series"].([]interface{})
	if !ok {
		t.Fatalf("cpuOption.series not a slice")
	}
	for _, s := range series {
		m, ok := s.(map[string]interface{})
		if !ok {
			t.Errorf("series element not a map: %#v", s)
			continue
		}
		dataArr, ok := m["data"].([]interface{})
		if !ok {
			t.Errorf("series data not a slice: %#v", m["data"])
			continue
		}
		if len(dataArr) != 0 {
			t.Errorf("expected empty data slices, got length %d", len(dataArr))
		}
	}
}

// TestGenerateReport_MultipleDevices ensures all device charts are rendered.
func TestGenerateReport_MultipleDevices(t *testing.T) {
	parsed := makeDummyParsedData()
	// Add a second device
	parsed.Devices["sdb"] = parsed.Devices["sda"]
	dir := t.TempDir()
	out := filepath.Join(dir, "multi.html")

	if err := GenerateReport(parsed, out, "Multi", "", "multi.log", "abcdef123456", "v0.0"); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	html := string(data)

	// Both device names appear
	if !strings.Contains(html, "sda") || !strings.Contains(html, "sdb") {
		t.Error("expected both sda and sdb in output")
	}

	// There should be one CPU chart, two device charts, and the emphasis call -> 4 setOption calls
	count := strings.Count(html, ".setOption(")
	if count != 4 {
		t.Errorf("expected 4 setOption calls, got %d", count)
	}
}

// TestGenerateReport_ShortHash does not abbreviate short hashes.
func TestGenerateReport_ShortHash(t *testing.T) {
	parsed := makeDummyParsedData()
	dir := t.TempDir()
	out := filepath.Join(dir, "short.html")

	if err := GenerateReport(parsed, out, "ShortHash", "", "f.log", "abcde", ""); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	html := string(data)

	// Should show full hash without truncation
	if !strings.Contains(html, "abcde") {
		t.Error("expected short hash abcde in output")
	}
	if strings.Contains(html, "abcde..") {
		t.Error("did not expect truncated hash")
	}
}

// TestGenerateReport_DeviceMBConversion verifies KB→MB conversion in device chart.
func TestGenerateReport_DeviceMBConversion(t *testing.T) {
	parsed := makeDummyParsedData()
	dir := t.TempDir()
	out := filepath.Join(dir, "mbconv.html")

	if err := GenerateReport(parsed, out, "MBConv", "", "mb.log", "hashhash", ""); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	html := string(data)

	// Extract the second setOption (first CPU, then device)
	parts := strings.Split(html, "c.setOption(")
	if len(parts) < 2 {
		t.Fatal("cannot locate device JSON")
	}
	raw := strings.SplitN(parts[1], ");", 2)[0]
	jsonTxt := strings.TrimSuffix(raw, "\n")
	var opt map[string]interface{}
	if err := json.Unmarshal([]byte(jsonTxt), &opt); err != nil {
		t.Fatalf("Device JSON invalid: %v", err)
	}
	series, ok := opt["series"].([]interface{})
	if !ok {
		t.Fatalf("option.series not a slice")
	}
	// The "Read MB/s" series is the third one (index 2)
	third, ok := series[2].(map[string]interface{})
	if !ok {
		t.Fatalf("third series not a map")
	}
	dataArr, ok := third["data"].([]interface{})
	if !ok {
		t.Fatalf("third series data not a slice")
	}
	// First sample: 1024KB → 1MB; second: 512KB → 0.5MB
	first, ok0 := dataArr[0].(float64)
	sec, ok1 := dataArr[1].(float64)
	if !ok0 || !ok1 {
		t.Fatalf("unexpected data types: %#v", dataArr)
	}
	if diff := math.Abs(first - 1.0); diff > 1e-9 {
		t.Errorf("expected first MB value ~1.0, got %v", first)
	}
	if diff := math.Abs(sec - 0.5); diff > 1e-9 {
		t.Errorf("expected second MB value ~0.5, got %v", sec)
	}
}
