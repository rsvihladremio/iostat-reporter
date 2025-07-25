# Iostat Reporter

A simple CLI application that outputs the result of `iostat -x -c -d -t 1 600` as an HTML report.


## How to use

Install is just provided by Go for the time being.

```bash
go install github.com/rsvihladremio/threaded-top-reporter@latest

# Or build from source using the Makefile:
```bash
make build     # compile the binary to bin/iorep
make lint      # run linters
make security  # run security checks
make fmt       # format code
make test      # run tests
make all       # run full build, lint, security, fmt, and test
```
```

Minimal usage is simple: provide an input file to generate a report with the default title `Iostat Report` and output to `iostat.html`.

```bash
iorep iostat.txt
report 'Iostat Report' written to iostat.html
```

Custom Report Output

```bash
iorep iostat.txt -o out.html
report 'Iostat Report' written to out.html
```

Custom Report Title

```bash
iorep iostat.txt -n 'My Report'
report 'My Report' written to iostat.html
```

Extra report metadata

```bash
iorep iostat.txt -m '{"id":"40de949f-3741-476a-abcb-3214a14ac15e"}'
report 'Iostat Report' written to iostat.html
```

## How it works

1. The arguments from the CLI are read such as the name (-n) of the report and the output location (optional but is -o)
2. The input file is read and parsed fully by the CLI
3. HTML is produced and written to disk at the output location (default is iostat.html).


### Metadata

Metadata is provided under the title as "details" atm there are no plans to allow more complex structures than key=value pairs.

### Charts

Charts are provided by [echarts](https://echarts.apache.org/), with the initial goal is to have the following:

* iowait, sys, user, idle, nice, steal usage plotted over time in a chart
* %time iowait was above 10% to show %time iobound the measurement was in a summary at the top
* %time user+system+nice+steal usage was above 50% to show CPU saturation for two thread per core machines
* %time user+system+nice+steal usage was above 90% to show CPU saturation for single thread per core machines
* %time each device had above 1.0 avg queue depth indicating the IO was saturated.
* p50 avg wait time for each device (put next to %time IO was saturated to show the effect)
* for each device do a chart for all disk IO stats in iostat
