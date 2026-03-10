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
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Jigsaw-Code/ech-research/internal/echtest"
	"github.com/Jigsaw-Code/ech-research/internal/soax"
	"github.com/Jigsaw-Code/ech-research/internal/workspace"
	"golang.org/x/sync/semaphore"
)

type TestResult struct {
	echtest.TestResult
	Country     string
	CountryName string
	ISP         string
	ASN         string
	ExitNodeIP  string
	ExitNodeISP string
}

func runSoaxTest(
	curlPath string,
	domain string,
	country string,
	countryName string,
	isp string,
	proxyURL string,
	echGrease bool,
	maxTime time.Duration,
) TestResult {
	headers := []string{"Respond-With: ip,isp,asn"}
	res := echtest.Run(curlPath, domain, echGrease, maxTime, proxyURL, headers)

	result := TestResult{
		TestResult:  res,
		Country:     country,
		CountryName: countryName,
		ISP:         isp,
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
			result.ExitNodeISP = val
		}
	}

	return result
}

type Country struct {
	Name string
	Code string
}

func loadCountries(path string) ([]Country, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var countries []Country
	reader := csv.NewReader(f)
	reader.Comment = '#' // Support skipping lines starting with #
	reader.FieldsPerRecord = 2

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read countries CSV: %w", err)
	}

	for i, record := range records {
		name := strings.TrimSpace(record[0])
		code := strings.TrimSpace(record[1])

		// Skip header row if present
		if i == 0 && strings.EqualFold(name, "Name") && strings.EqualFold(code, "Code") {
			continue
		}

		countries = append(countries, Country{
			Name: name,
			Code: code,
		})
	}
	return countries, nil
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
		parallelismFlag  = flag.Int("parallelism", 16, "Maximum number of parallel requests")
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
	countriesPath := *countriesFlag
	if countriesPath == "" {
		countriesPath = filepath.Join(workspaceDir, "countries.csv")
	}
	countries, err := loadCountries(countriesPath)
	if err != nil {
		slog.Error("Failed to load countries list", "path", countriesPath, "error", err)
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
			"domain", "country_code", "country_name", "isp", "asn", "exit_node_ip", "exit_node_isp", "ech_grease", "error",
			"curl_exit_code", "curl_error_name", "dns_lookup_ms", "tcp_connection_ms",
			"tls_handshake_ms", "server_time_ms", "total_time_ms", "http_status",
		}
		if err := csvWriter.Write(header); err != nil {
			slog.Error("Failed to write CSV header", "error", err)
		}

		for r := range resultsCh {
			record := []string{
				r.Domain, r.Country, r.CountryName, r.ISP, r.ASN, r.ExitNodeIP, r.ExitNodeISP, strconv.FormatBool(r.ECHGrease), r.Error,
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
	runSessionID := time.Now().Format("0102150405")
	sem := semaphore.NewWeighted(int64(*parallelismFlag))
	var wg sync.WaitGroup
	var total, finished atomic.Int32

	// Audit map to store discovered ISPs per country
	ispAuditMap := make(map[string][]string)

	for _, country := range countries {
		slog.Debug("Processing country", "name", country.Name, "code", country.Code)

		isps, err := client.ListISPs(country.Code)
		if err != nil {
			slog.Error("Failed to fetch ISPs", "country", country.Code, "error", err)
			continue
		}

		ispAuditMap[country.Code] = isps

		total.Add(int32(len(isps) * 2))
		for i, isp := range isps {
			wg.Add(2)
			sessionID := fmt.Sprintf("%s%s%d", runSessionID, country.Code, i)

			startTest := func(c Country, isp, sid string, ech bool) {
				defer wg.Done()
				if err := sem.Acquire(context.Background(), 1); err != nil {
					slog.Error("Failed to acquire semaphore", "error", err)
					return
				}
				defer sem.Release(1)

				proxyURL := client.BuildWebProxyURL(c.Code, isp, sid)
				slog.Debug("Testing ISP", "country", c.Code, "isp", isp, "ech_grease", ech, "session", sid)
				resultsCh <- runSoaxTest(curlPath, domain, c.Code, c.Name, isp, proxyURL, ech, *maxTimeFlag)
				progress := fmt.Sprintf("%d/%d", finished.Add(1), total.Load())
				slog.Info("Finished", "country", c.Code, "isp", isp, "progress", progress)
			}

			go startTest(country, isp, sessionID, false)
			go startTest(country, isp, sessionID, true)
		}
	}

	wg.Wait()
	close(resultsCh)
	csvWg.Wait()

	// Write the ISP audit log to JSON
	auditFilename := filepath.Join(workspaceDir, "soax-isps-audit.json")
	auditData, err := json.MarshalIndent(ispAuditMap, "", "  ")
	if err == nil {
		if err := os.WriteFile(auditFilename, auditData, 0644); err != nil {
			slog.Error("Failed to write ISP audit log", "error", err)
		} else {
			slog.Info("ISP audit log saved", "path", auditFilename)
		}
	} else {
		slog.Error("Failed to marshal ISP audit log", "error", err)
	}

	slog.Info("Done. Results saved to", "path", outputFilename)
}
