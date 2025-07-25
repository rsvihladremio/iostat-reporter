package reporter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/rsvihladremio/iostat-reporter/parser"
)

func GenerateReport(parsedData parser.ParsedData, outputFile string, reportTitle string, metadata string, fileName string, fileHash string, version string) error {
	// Build time axis and CPU series
	times := make([]string, len(parsedData.CPUs))
	users := make([]float64, len(parsedData.CPUs))
	systems := make([]float64, len(parsedData.CPUs))
	idles := make([]float64, len(parsedData.CPUs))
	iowaits := make([]float64, len(parsedData.CPUs))
	nices := make([]float64, len(parsedData.CPUs))
	steals := make([]float64, len(parsedData.CPUs))

	for i, cs := range parsedData.CPUs {
		times[i] = cs.Timestamp.Format("15:04:05")
		users[i] = cs.User
		systems[i] = cs.System
		idles[i] = cs.Idle
		iowaits[i] = cs.Iowait
		nices[i] = cs.Nice
		steals[i] = cs.Steal
	}

	cpuOption := map[string]interface{}{
		"tooltip": map[string]interface{}{"trigger": "axis"},
		"legend":  map[string]interface{}{"data": []string{"User", "System", "Idle", "IOWait", "Nice", "Steal"}, "bottom": 0},
		"toolbox": map[string]interface{}{
			"show": true,
			"top":  -7,
			"feature": map[string]interface{}{
				"saveAsImage": map[string]interface{}{},
				"dataZoom":    map[string]interface{}{},
				"dataView":    map[string]interface{}{"readOnly": false},
				"restore":     map[string]interface{}{},
			},
		},
		"xAxis": map[string]interface{}{"type": "category", "data": times},
		"yAxis": map[string]interface{}{"type": "value", "name": "% CPU"},
		"series": []map[string]interface{}{
			{"name": "User", "type": "line", "data": users},
			{"name": "System", "type": "line", "data": systems},
			{"name": "Idle", "type": "line", "data": idles},
			{"name": "IOWait", "type": "line", "data": iowaits},
			{"name": "Nice", "type": "line", "data": nices},
			{"name": "Steal", "type": "line", "data": steals},
		},
	}

	cpuJSON, err := json.Marshal(cpuOption)
	if err != nil {
		return fmt.Errorf("failed to marshal CPU options: %w", err)
	}

	// Build per-device charts
	type DeviceChart struct {
		DeviceName string
		ChartID    string
		OptionJSON template.JS
	}

	var deviceCharts []DeviceChart
	for dev, stats := range parsedData.Devices {
		reqReads := make([]float64, len(stats))
		reqWrites := make([]float64, len(stats))
		kbReads := make([]float64, len(stats))
		kbWrites := make([]float64, len(stats))
		latReads := make([]float64, len(stats))
		latWrites := make([]float64, len(stats))
		queueSizes := make([]float64, len(stats))
		for i, ds := range stats {
			reqReads[i] = ds.ReadsPerSec
			reqWrites[i] = ds.WritesPerSec
			kbReads[i] = ds.ReadKBPerSec / 1024.0
			kbWrites[i] = ds.WriteKBPerSec / 1024.0
			latReads[i] = ds.ReadAwaitMs
			latWrites[i] = ds.WriteAwaitMs
			queueSizes[i] = ds.QueueSize
		}
		const numSplits = 5
		// Compute min, max, and interval for each axis group with 5 splits
		reqMin, reqMax, reqInterval := CalcScale(numSplits, reqReads, reqWrites)
		kbMin, kbMax, kbInterval := CalcScale(numSplits, kbReads, kbWrites)
		latMin, latMax, latInterval := CalcScale(numSplits, latReads, latWrites)
		qMin, qMax, qInterval := CalcScale(numSplits, queueSizes)
		chartID := "dev_" + strings.ReplaceAll(dev, "-", "_") + "_chart"
		option := map[string]interface{}{
			"grid":    map[string]interface{}{"containLabel": true},
			"tooltip": map[string]interface{}{"trigger": "axis"},
			"legend": map[string]interface{}{
				"data":   []string{"Read Req/s", "Write Req/s", "Read MB/s", "Write MB/s", "Read Latency (ms)", "Write Latency (ms)", "Queue Size"},
				"bottom": 0,
			},
			"toolbox": map[string]interface{}{
				"show": true,
				"top":  -7,
				"feature": map[string]interface{}{
					"saveAsImage": map[string]interface{}{},
					"dataZoom":    map[string]interface{}{},
					"dataView":    map[string]interface{}{"readOnly": false},
					"restore":     map[string]interface{}{},
				},
			},
			"xAxis": map[string]interface{}{"type": "category", "data": times},
			"yAxis": []map[string]interface{}{
				{
					"type":      "value",
					"name":      "Req/s",
					"position":  "left",
					"min":       reqMin,
					"max":       reqMax,
					"interval":  reqInterval,
					"splitLine": map[string]interface{}{"show": true},
					"axisLabel": map[string]interface{}{"formatter": "{value} req/s"},
				},
				{
					"type":      "value",
					"name":      "MB/s",
					"position":  "left",
					"offset":    80,
					"min":       kbMin,
					"max":       kbMax,
					"interval":  kbInterval,
					"splitLine": map[string]interface{}{"show": true},
					"axisLabel": map[string]interface{}{"formatter": "{value} MB/s"},
				},
				{
					"type":      "value",
					"name":      "ms",
					"position":  "right",
					"offset":    40,
					"min":       latMin,
					"max":       latMax,
					"interval":  latInterval,
					"splitLine": map[string]interface{}{"show": true},
					"axisLabel": map[string]interface{}{"formatter": "{value} ms"},
				},
				{
					"type":      "value",
					"name":      "Queue Size",
					"position":  "right",
					"offset":    120,
					"min":       qMin,
					"max":       qMax,
					"interval":  qInterval,
					"splitLine": map[string]interface{}{"show": true},
					"axisLabel": map[string]interface{}{"formatter": "{value}"},
				},
			},
			"series": []map[string]interface{}{
				{"name": "Read Req/s", "type": "line", "data": reqReads, "yAxisIndex": 0},
				{"name": "Write Req/s", "type": "line", "data": reqWrites, "yAxisIndex": 0},
				{"name": "Read MB/s", "type": "line", "data": kbReads, "yAxisIndex": 1},
				{"name": "Write MB/s", "type": "line", "data": kbWrites, "yAxisIndex": 1},
				{"name": "Read Latency (ms)", "type": "line", "data": latReads, "yAxisIndex": 2},
				{"name": "Write Latency (ms)", "type": "line", "data": latWrites, "yAxisIndex": 2},
				{"name": "Queue Size", "type": "line", "data": queueSizes, "yAxisIndex": 3},
			},
		}

		js, err := json.Marshal(option)
		if err != nil {
			return fmt.Errorf("failed to marshal device chart for %s: %w", dev, err)
		}

		deviceCharts = append(deviceCharts, DeviceChart{
			DeviceName: dev,
			ChartID:    chartID,
			OptionJSON: template.JS(js), // #nosec G203
		})
	}

	// Render the template
	templatePath := "templates/report.html"
	if _, statErr := os.Stat(templatePath); os.IsNotExist(statErr) {
		templatePath = filepath.Join("..", "templates", "report.html")
	}
	tmpl, err := template.New("report.html").Funcs(template.FuncMap{
		"safeJS": func(s string) template.JS { return template.JS(s) }, // #nosec G203
		"abbr": func(s string) string {
			if len(s) > 6 {
				return s[:6] + ".."
			}
			return s
		},
	}).ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %q: %w", templatePath, err)
	}

	data := struct {
		Title        string
		Metadata     string
		FileName     string
		FileHash     string
		Version      string
		CpuOption    template.JS
		DeviceCharts []DeviceChart
	}{
		Title:        reportTitle,
		Metadata:     metadata,
		FileName:     fileName,
		FileHash:     fileHash,
		Version:      version,
		CpuOption:    template.JS(cpuJSON), // #nosec G203
		DeviceCharts: deviceCharts,
	}

	f, err := os.Create(outputFile) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close output file: %v\n", cerr)
		}
	}()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
