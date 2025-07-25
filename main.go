package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/spf13/pflag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rsvihladremio/iostat-reporter/parser" // Import the parser package
	"github.com/rsvihladremio/iostat-reporter/reporter"
)

var (
	outputFile  string
	reportTitle string
	metadata    string
	Version     string = "dev" // overridden via -ldflags "-X main.Version=…"
)

func init() {
	pflag.StringVarP(&outputFile, "output", "o", "iostat.html", "Output HTML file path")
	pflag.StringVarP(&reportTitle, "name", "n", "Iostat Report", "Report title")
	pflag.StringVarP(&metadata, "metadata", "m", "", "Additional metadata as JSON string")
	showVersion := pflag.Bool("version", false, "show version and exit")

	pflag.Parse()

	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}
}

func main() {

	// Validate input file
	args := pflag.Args()
	if len(args) < 1 {
		log.Fatal("Please provide an input file")
	}
	inputFile := args[0]

	// Sanitize and read input file
	cleanInput := filepath.Clean(inputFile)
	if strings.Contains(cleanInput, "..") {
		log.Fatalf("invalid input path: %s", inputFile)
	}
	data, err := os.ReadFile(cleanInput)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Parse input using the new package
	parsedData, err := parser.ParseIostatOutput(data)
	if err != nil {
		log.Fatalf("Error parsing top output: %v", err)
	}

	// Generate report
	fileName := filepath.Base(cleanInput)
	fileHash := fmt.Sprintf("%x", sha256.Sum256(data))
	if err := reporter.GenerateReport(parsedData, outputFile, reportTitle, metadata, fileName, fileHash, Version); err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	fmt.Printf("report '%s' written to %s\n", reportTitle, outputFile)
}
