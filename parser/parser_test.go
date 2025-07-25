package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseTimestamp verifies correct parsing of timestamp lines.
func TestParseTimestamp(t *testing.T) {
	ts, err := parseTimestamp("09/04/24 12:07:20")
	assert.NoError(t, err)
	// Compare only formatted time to avoid timezone issues.
	assert.Equal(t, "2024-09-04 12:07:20", ts.Format("2006-01-02 15:04:05"))
}

// TestParseCPUStats checks that CPU fields are mapped correctly.
func TestParseCPUStats(t *testing.T) {
	baseTs := time.Now()
	fields := []string{"2.36", "0.00", "0.40", "0.04", "0.01", "97.20"}
	cpu, err := parseCPUStats(baseTs, fields)
	assert.NoError(t, err)
	assert.Equal(t, 2.36, cpu.User)
	assert.Equal(t, 0.40, cpu.System)
	assert.Equal(t, 97.20, cpu.Idle)
	assert.Equal(t, baseTs, cpu.Timestamp)
}

// TestParseDeviceStats checks that device fields are mapped correctly.
func TestParseDeviceStats(t *testing.T) {
	baseTs := time.Now()
	fields := []string{
		"sda", "2.08", "94.38", "0.31", "13.07", "0.89", "45.47",
		"9.58", "210.39", "5.55", "36.68", "2.74", "21.96",
	}
	dev, err := parseDeviceStats(baseTs, fields)
	assert.NoError(t, err)
	assert.Equal(t, "sda", dev.Name)
	assert.Equal(t, 2.08, dev.ReadsPerSec)
	assert.Equal(t, 21.96, dev.WriteReqSzKB)
	assert.Equal(t, baseTs, dev.Timestamp)
}

// TestParseIostatOutput does a small integration test on a sample block.
func TestParseIostatOutput(t *testing.T) {
	sample := `
09/04/24 12:07:20
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
           2.36    0.00    0.40    0.04    0.01   97.20

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz
sda              2.08     94.38     0.31  13.07    0.89    45.47    9.58    210.39     5.55  36.68    2.74    21.96
`
	p, err := ParseIostatOutput([]byte(sample))
	assert.NoError(t, err)

	// one CPU record
	assert.Len(t, p.CPUs, 1)
	assert.Equal(t, 2.36, p.CPUs[0].User)
	assert.Equal(t, 97.20, p.CPUs[0].Idle)

	// one device record under "sda"
	recs, ok := p.Devices["sda"]
	assert.True(t, ok)
	assert.Len(t, recs, 1)
	assert.Equal(t, 9.58, recs[0].WritesPerSec)
}

// TestParseIostatOutputExtended verifies parsing of iostat output with extended device columns.
func TestParseIostatOutputExtended(t *testing.T) {
	sample := `
09/04/24 12:07:26
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          21.73    0.00    5.93    0.49    0.25   71.60

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.00      0.00     0.00   0.00    0.00     0.00  161.39    550.50    52.48  24.54    0.35     3.41    0.00      0.00     0.00   0.00    0.00     0.00   82.18    0.06    0.06  19.41
`
	p, err := ParseIostatOutput([]byte(sample))
	assert.NoError(t, err)

	// one CPU record
	assert.Len(t, p.CPUs, 1)
	assert.Equal(t, 21.73, p.CPUs[0].User)
	assert.Equal(t, 71.60, p.CPUs[0].Idle)

	// one device record under "sda"
	recs, ok := p.Devices["sda"]
	assert.True(t, ok)
	assert.Len(t, recs, 1)
	assert.Equal(t, 161.39, recs[0].WritesPerSec)
	assert.Equal(t, 3.41, recs[0].WriteReqSzKB)
}

func TestParseIostatFullSample(t *testing.T) {
	const fullSample = `
09/04/24 12:07:20
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
           2.36    0.00    0.40    0.04    0.01   97.20

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              2.08     94.38     0.31  13.07    0.89    45.47    9.58    210.39     5.55  36.68    2.74    21.96    0.09    377.20     0.00   0.00    0.95  4151.86    3.94    0.06    0.03   1.39

09/04/24 12:07:21
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          33.91    0.00    7.67    2.72    0.00   55.69

Device            r/s   rkB/s   rrqm/s  %rrqm r_await rareq-sz    w/s    wkB/s  wrqm/s  %wrqm w_await wareq-sz    d/s   dkB/s  drqm/s  %drqm d_await dareq-sz  f/s f_await aqu-sz %util
sda              0.00    0.00     0.00   0.00    0.00     0.00  395.00 38116.00 133.00  25.19    8.65     96.50     1.00   4.00   0.00    0.00    1.00     4.00 122.00   0.06   3.42 39.20

09/04/24 12:07:22
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          26.76    0.00    4.38    0.73    0.24   67.88

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.00      0.00     0.00   0.00    0.00     0.00  201.00    792.00    90.00  30.93    0.35     3.94    0.00      0.00     0.00   0.00    0.00     0.00  103.00    0.07    0.08  26.80

09/04/24 12:07:23
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          33.25    0.00    7.94    0.74    0.00   58.06

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.00      0.00     0.00   0.00    0.00     0.00  282.00   1128.00   130.00  31.55    0.35     4.00    1.00      4.00     0.00   0.00    0.00     4.00  144.00    0.06    0.11  34.00

09/04/24 12:07:24
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          24.81    0.00    5.21    0.50    0.00   69.48

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.00      0.00     0.00   0.00    0.00     0.00  198.00    720.00    75.00  27.47    0.33     3.64    0.00      0.00     0.00   0.00    0.00     0.00  101.00    0.06    0.07  24.80

09/04/24 12:07:25
avg-cpu:  %user   %nice %system %iowait  %steal   %idle
          17.53    0.00    5.43    0.25    0.00   76.79

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.00      0.00     0.00   0.00    0.00     0.00  231.00   2688.00   151.00  39.53    0.80    11.64    0.00      0.00     0.00   0.00    0.00     0.00   72.00    0.07    0.19  19.60

09/04/24 12:07:26
avg-cpu:  %user   %nice %system %iowait  %steal   **idle**
          21.73    0.00    5.93    0.49    0.25   71.60

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.00      0.00     0.00   0.00    0.00     0.00  161.39    550.50    52.48  24.54    0.35     3.41    0.00      0.00     0.00   0.00    0.00     0.00   82.18    0.06    0.06  19.41
`

	parsed, err := ParseIostatOutput([]byte(fullSample))
	assert.NoError(t, err)

	// 1) Expect one CPUStats per timestamp
	assert.Len(t, parsed.CPUs, 7, "should parse all 7 CPU frames")

	// Spot-check first CPU frame:
	first := parsed.CPUs[0]
	assert.Equal(t, "2024-09-04 12:07:20", first.Timestamp.Format("2006-01-02 15:04:05"))
	assert.InDelta(t, 2.36, first.User, 1e-6)
	assert.InDelta(t, 97.20, first.Idle, 1e-6)

	// Spot-check middle CPU frame (index 3)
	mid := parsed.CPUs[3]
	assert.InDelta(t, 33.25, mid.User, 1e-6)
	assert.InDelta(t, 58.06, mid.Idle, 1e-6)

	// Expect exactly 7 DeviceStats for "sda"
	sdaRecs, ok := parsed.Devices["sda"]
	assert.True(t, ok, "should have stats for sda")
	assert.Len(t, sdaRecs, 7)

	// Spot-check first and last device entries:
	d0 := sdaRecs[0]
	assert.Equal(t, "sda", d0.Name)
	assert.InDelta(t, 2.08, d0.ReadsPerSec, 1e-6)
	assert.InDelta(t, 21.96, d0.WriteReqSzKB, 1e-6)

	dLast := sdaRecs[6]
	assert.Equal(t, "2024-09-04 12:07:26", dLast.Timestamp.Format("2006-01-02 15:04:05"))
	assert.InDelta(t, 161.39, dLast.WritesPerSec, 1e-6)
	assert.InDelta(t, 3.41, dLast.WriteReqSzKB, 1e-6)
}

func TestParseIostatMultipleDevices(t *testing.T) {
	const multiDevSample = `
09/04/24 12:10:00
avg-cpu:  %user   %nice  %system  %iowait   %steal   %idle
           5.00     0.00     1.00     0.10     0.00   93.90

Device            r/s     rkB/s   rrqm/s  %rrqm  r_await  rareq-sz   w/s     wkB/s   wrqm/s  %wrqm  w_await  wareq-sz
sda               1.23     50.00     0.10   8.33     2.00     40.65   10.00    100.00    1.00  9.09     1.50     10.00
sdb               9.87    200.00     0.00   0.00     1.00     20.00    5.00     75.00    0.50  9.09     0.80      15.00
`

	report, err := ParseIostatOutput([]byte(multiDevSample))
	assert.NoError(t, err)

	// one CPU snapshot
	assert.Len(t, report.CPUs, 1)
	cpu := report.CPUs[0]
	assert.InDelta(t, 5.00, cpu.User, 1e-6)
	assert.InDelta(t, 93.90, cpu.Idle, 1e-6)

	// two devices present
	assert.Len(t, report.Devices, 2, "should see two distinct device keys")

	// verify sda
	sdaRecs, ok := report.Devices["sda"]
	assert.True(t, ok, "entry for sda should exist")
	assert.Len(t, sdaRecs, 1, "one sample for sda")
	assert.InDelta(t, 1.23, sdaRecs[0].ReadsPerSec, 1e-6)
	assert.InDelta(t, 10.00, sdaRecs[0].WritesPerSec, 1e-6)

	// verify sdb
	sdbRecs, ok := report.Devices["sdb"]
	assert.True(t, ok, "entry for sdb should exist")
	assert.Len(t, sdbRecs, 1, "one sample for sdb")
	assert.InDelta(t, 9.87, sdbRecs[0].ReadsPerSec, 1e-6)
	assert.InDelta(t, 5.00, sdbRecs[0].WritesPerSec, 1e-6)
}
