package reporter

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateReport_DeviceAxisScales(t *testing.T) {
	parsed := makeDummyParsedData()
	dir := t.TempDir()
	out := filepath.Join(dir, "axis.html")

	if err := GenerateReport(parsed, out, "AxisScales", "", "axis.log", "hashhash", ""); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	html := string(data)
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
	yAxes, ok := opt["yAxis"].([]interface{})
	if !ok {
		t.Fatalf("expected yAxis to be a slice, got %T", opt["yAxis"])
	}
	if len(yAxes) != 4 {
		t.Fatalf("expected 4 yAxis entries, got %d", len(yAxes))
	}

	// Axis 0: Req/s -> [1,3,2,4] => min=1, max=4, interval=0.6
	axis0 := yAxes[0].(map[string]interface{})
	min0 := axis0["min"].(float64)
	max0 := axis0["max"].(float64)
	int0 := axis0["interval"].(float64)
	if min0 != 1 {
		t.Errorf("axis0 min = %v; want 1", min0)
	}
	if max0 != 4 {
		t.Errorf("axis0 max = %v; want 4", max0)
	}
	if diff := math.Abs(int0 - 0.6); diff > 1e-9 {
		t.Errorf("axis0 interval = %v; want ~0.6", int0)
	}

	// Axis 1: MB/s -> [1,0.5,2,1] => min=0.5, max=2, interval=0.3
	axis1 := yAxes[1].(map[string]interface{})
	min1 := axis1["min"].(float64)
	max1 := axis1["max"].(float64)
	int1 := axis1["interval"].(float64)
	if diff := math.Abs(min1 - 0.5); diff > 1e-9 {
		t.Errorf("axis1 min = %v; want 0.5", min1)
	}
	if max1 != 2 {
		t.Errorf("axis1 max = %v; want 2", max1)
	}
	if diff := math.Abs(int1 - 0.3); diff > 1e-9 {
		t.Errorf("axis1 interval = %v; want ~0.3", int1)
	}

	// Axis 2: ms -> [0.5,0.2,1.5,0.8] => min=0.2, max=1.5, interval=0.26
	axis2 := yAxes[2].(map[string]interface{})
	min2 := axis2["min"].(float64)
	max2 := axis2["max"].(float64)
	int2 := axis2["interval"].(float64)
	if diff := math.Abs(min2 - 0.2); diff > 1e-9 {
		t.Errorf("axis2 min = %v; want 0.2", min2)
	}
	if diff := math.Abs(max2 - 1.5); diff > 1e-9 {
		t.Errorf("axis2 max = %v; want 1.5", max2)
	}
	if diff := math.Abs(int2 - 0.26); diff > 1e-9 {
		t.Errorf("axis2 interval = %v; want ~0.26", int2)
	}

	// Axis 3: Queue Size -> [0.1,0.05] => min=0.05, max=0.1, interval=0.01
	axis3 := yAxes[3].(map[string]interface{})
	min3 := axis3["min"].(float64)
	max3 := axis3["max"].(float64)
	int3 := axis3["interval"].(float64)
	if diff := math.Abs(min3 - 0.05); diff > 1e-9 {
		t.Errorf("axis3 min = %v; want 0.05", min3)
	}
	if diff := math.Abs(max3 - 0.1); diff > 1e-9 {
		t.Errorf("axis3 max = %v; want 0.1", max3)
	}
	if diff := math.Abs(int3 - 0.01); diff > 1e-9 {
		t.Errorf("axis3 interval = %v; want ~0.01", int3)
	}
}
