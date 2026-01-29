// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jigsaw-Code/ech-research/internal/curl"
	"github.com/Jigsaw-Code/ech-research/internal/soax"
	"github.com/Jigsaw-Code/ech-research/internal/workspace"
	"golang.org/x/sync/semaphore"
)

type TestResult struct {
	Domain        string
	Country       string
	ISP           string
	ASN           string
	ExitNodeIP    string
	ECHGrease     bool
	Error         string
	CurlExitCode  int
	CurlErrorName string
	DNSLookup     time.Duration
	TCPConnection time.Duration
	TLSHandshake  time.Duration
	ServerTime    time.Duration
	TotalTime     time.Duration
	HTTPStatus    int
}

func runSoaxTest(
	runner *curl.Runner,
	domain string,
	country string,
	isp string,
	proxyURL string,
	echGrease bool,
	maxTime time.Duration,
) TestResult {
	result := TestResult{
		Domain:    domain,
		Country:   country,
		ISP:       isp,
		ECHGrease: echGrease,
	}

	echMode := curl.ECHFalse
	if echGrease {
		echMode = curl.ECHGrease
	}

	url := "https://" + domain
	res, err := runner.Run(url, curl.Args{
		Proxy:        proxyURL,
		ProxyHeaders: []string{"Respond-With: ip,isp,asn"},
		ECH:          echMode,
		Timeout:      maxTime,
		Verbose:      true, // Required to capture response headers
		MeasureStats: true,
	})

	result.CurlExitCode = res.ExitCode
	result.CurlErrorName = curl.ExitCodeName(res.ExitCode)
	result.HTTPStatus = res.Stats.HTTPStatus
	result.DNSLookup = res.Stats.DNSLookupTimestamp
	result.TCPConnection = res.Stats.TCPConnectTimestamp
	result.TLSHandshake = res.Stats.TLSConnectTimestamp
	result.ServerTime = res.Stats.ServerResponseTimestamp
	result.TotalTime = res.Stats.TotalTimeTimestamp

	if err != nil {
		result.Error = err.Error()
	}

	// Parse metadata from Stderr (SOAX specific headers in CONNECT response)
	for line := range strings.SplitSeq(res.Stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "< Node-") {
			continue
		}
		kv := strings.SplitN(strings.TrimPrefix(line, "< Node-"), ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])
		switch key {
		case "asn":
			result.ASN = val
		case "ip":
			result.ExitNodeIP = val
		case "isp":
			result.ISP += " (" + val + ")"
		}
	}

	return result
}

func loadCountries(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var countries []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			countries = append(countries, line)
		}
	}
	return countries, scanner.Err()
}

func main() {
	var (
		workspaceFlag    = flag.String("workspace", "./workspace", "Directory to store intermediate files")
		soaxConfigFlag   = flag.String("soax", "", "Path to SOAX config JSON")
		countriesFlag    = flag.String("countries", "", "Path to file containing ISO country codes")
		targetDomainFlag = flag.String("targetDomain", "www.google.com", "Target domain to test")
		verboseFlag      = flag.Bool("verbose", false, "Enable verbose logging")
		maxTimeFlag      = flag.Duration("maxTime", 30*time.Second, "Maximum time per curl request")
		curlPathFlag     = flag.String("curl", "", "Path to the ECH-enabled curl binary")
		parallelismFlag  = flag.Int("parallelism", 10, "Maximum number of parallel requests")
	)
	flag.Parse()

	if *verboseFlag {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	}

	// Set up workspace directory
	workspaceDir := workspace.EnsureWorkspace(*workspaceFlag)

	// Determine curl binary path
	curlPath := *curlPathFlag
	if curlPath == "" {
		curlPath = filepath.Join(workspaceDir, "output", "bin", "curl")
	}
	runner := curl.NewRunner(curlPath)

	// Load SOAX config
	soaxConfigPath := *soaxConfigFlag
	if soaxConfigPath == "" {
		soaxConfigPath = filepath.Join(workspaceDir, "soax", "cred.json")
	}
	cfg, err := soax.LoadConfig(soaxConfigPath)
	if err != nil {
		slog.Error("Failed to load SOAX config", "path", soaxConfigPath, "error", err)
		os.Exit(1)
	}
	client := soax.NewClient(cfg)

	// Load countries
	if *countriesFlag == "" {
		slog.Error("The --countries flag is required")
		os.Exit(1)
	}
	countries, err := loadCountries(*countriesFlag)
	if err != nil {
		slog.Error("Failed to load countries list", "path", *countriesFlag, "error", err)
		os.Exit(1)
	}

	// Create output CSV file
	sanitizedDomain := strings.ReplaceAll(*targetDomainFlag, ".", "_")
	outputFilename := filepath.Join(workspaceDir, fmt.Sprintf("soax-results-%s-countries%d.csv", sanitizedDomain, len(countries)))
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		slog.Error("Failed to create output CSV file", "path", outputFilename, "error", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	resultsCh := make(chan TestResult, 2*len(countries)*(*parallelismFlag))

	var csvWg sync.WaitGroup
	csvWg.Add(1)
	go func() {
		defer csvWg.Done()
		csvWriter := csv.NewWriter(outputFile)
		defer csvWriter.Flush()

		header := []string{
			"domain", "country", "isp", "asn", "exit_node_ip", "ech_grease", "error",
			"curl_exit_code", "curl_error_name", "dns_lookup_ms", "tcp_connection_ms",
			"tls_handshake_ms", "server_time_ms", "total_time_ms", "http_status",
		}
		if err := csvWriter.Write(header); err != nil {
			slog.Error("Failed to write CSV header", "error", err)
		}

		for r := range resultsCh {
			record := []string{
				r.Domain, r.Country, r.ISP, r.ASN, r.ExitNodeIP, strconv.FormatBool(r.ECHGrease), r.Error,
				strconv.Itoa(r.CurlExitCode), r.CurlErrorName,
				strconv.FormatInt(r.DNSLookup.Milliseconds(), 10),
				strconv.FormatInt(r.TCPConnection.Milliseconds(), 10),
				strconv.FormatInt(r.TLSHandshake.Milliseconds(), 10),
				strconv.FormatInt(r.ServerTime.Milliseconds(), 10),
				strconv.FormatInt(r.TotalTime.Milliseconds(), 10),
				strconv.Itoa(r.HTTPStatus),
			}
			if err := csvWriter.Write(record); err != nil {
				slog.Error("Failed to write record to CSV", "error", err)
			}
		}
	}()

	domain := *targetDomainFlag
	sem := semaphore.NewWeighted(int64(*parallelismFlag))
	var wg sync.WaitGroup
	for _, country := range countries {
		slog.Debug("Processing country", "country", country)

		isps, err := client.ListISPs(country)
		if err != nil {
			slog.Error("Failed to fetch ISPs", "country", country, "error", err)
			continue
		}

		for _, isp := range isps {
			wg.Add(2)

			if err := sem.Acquire(context.Background(), 1); err != nil {
				slog.Error("Failed to acquire semaphore", "error", err)
				wg.Done()
			} else {
				go func(c, isp string) {
					defer sem.Release(1)
					defer wg.Done()
					proxyURL := client.BuildProxyURL(c, isp, "")
					slog.Info("Testing ISP", "country", c, "isp", isp, "ech_grease", false)
					resultsCh <- runSoaxTest(runner, domain, c, isp, proxyURL, false, *maxTimeFlag)
				}(country, isp)
			}

			if err := sem.Acquire(context.Background(), 1); err != nil {
				slog.Error("Failed to acquire semaphore", "error", err)
				wg.Done()
			} else {
				go func(c, isp string) {
					defer sem.Release(1)
					defer wg.Done()
					proxyURL := client.BuildProxyURL(c, isp, "")
					slog.Info("Testing ISP", "country", c, "isp", isp, "ech_grease", true)
					resultsCh <- runSoaxTest(runner, domain, c, isp, proxyURL, true, *maxTimeFlag)
				}(country, isp)
			}
		}
	}

	wg.Wait()
	close(resultsCh)
	csvWg.Wait()

	slog.Info("Done. Results saved to", "path", outputFilename)
}
