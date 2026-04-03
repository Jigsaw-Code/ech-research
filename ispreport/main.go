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
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Jigsaw-Code/ech-research/internal/echtest"
	"github.com/Jigsaw-Code/ech-research/internal/soax"
	"github.com/Jigsaw-Code/ech-research/internal/workspace"
	"github.com/oschwald/maxminddb-golang"
	"golang.org/x/sync/semaphore"
)

type TestResult struct {
	echtest.TestResult
	Country      string
	CountryName  string
	ISP          string
	ASN          string
	ExitNodeIP   string
	ExitNodeISP  string
	DiscoveredIP string
	IPMatch      string
	GeoDBASN     string
	GeoDBASName  string
	ASNMatch     string
}

func lookupASN(db *maxminddb.Reader, ipStr string) (string, string) {
	if db == nil || ipStr == "" {
		return "", ""
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", ""
	}

	var record struct {
		AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
		AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
	}

	err := db.Lookup(ip, &record)
	if err != nil {
		return "", ""
	}

	asn := ""
	if record.AutonomousSystemNumber > 0 {
		asn = fmt.Sprintf("%d", record.AutonomousSystemNumber)
	}
	return asn, record.AutonomousSystemOrganization
}

func discoverIP(curlPath, proxyURL string, maxTime time.Duration, discoveryURL string) string {
	args := []string{
		"-s",
		"-4",
		"--max-time", strconv.FormatFloat(maxTime.Seconds(), 'f', -1, 64),
		"--proxy", proxyURL,
		discoveryURL,
	}
	cmd := exec.Command(curlPath, args...)

	binDir := filepath.Dir(curlPath)
	libDir := filepath.Join(filepath.Dir(binDir), "lib")
	if libStat, err := os.Stat(libDir); err == nil && libStat.IsDir() {
		cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+libDir)
	}

	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
	ipCheckURL string,
	asnDB *maxminddb.Reader,
) TestResult {
	discoveredIP := discoverIP(curlPath, proxyURL, maxTime, ipCheckURL)

	headers := []string{"Respond-With: ip,isp,asn"}
	res := echtest.Run(curlPath, domain, echGrease, maxTime, proxyURL, headers)

	result := TestResult{
		TestResult:   res,
		Country:      country,
		CountryName:  countryName,
		ISP:          isp,
		DiscoveredIP: discoveredIP,
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

	if result.ExitNodeIP != "" && result.DiscoveredIP != "" {
		if result.ExitNodeIP == result.DiscoveredIP {
			result.IPMatch = "true"
		} else {
			result.IPMatch = "false"
		}
	}

	if asnDB != nil && result.DiscoveredIP != "" {
		result.GeoDBASN, result.GeoDBASName = lookupASN(asnDB, result.DiscoveredIP)
		if result.ASN != "" && result.GeoDBASN != "" {
			soaxASN := strings.TrimPrefix(strings.ToUpper(result.ASN), "AS")
			if soaxASN == result.GeoDBASN {
				result.ASNMatch = "true"
			} else {
				result.ASNMatch = "false"
			}
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
		countriesFlag    = flag.String("countries", "", "Path to file containing ISO country codes")
		targetDomainFlag = flag.String("targetDomain", "www.google.com", "Target domain to test")
		verboseFlag      = flag.Bool("verbose", false, "Enable verbose logging")
		maxTimeFlag      = flag.Duration("maxTime", 30*time.Second, "Maximum time per curl request")
		curlPathFlag     = flag.String("curl", "", "Path to the ECH-enabled curl binary")
		parallelismFlag  = flag.Int("parallelism", 16, "Maximum number of parallel requests")
		ipCheckURLFlag   = flag.String("ipCheckURL", "https://ipv4.icanhazip.com/", "URL for checking the real exit IP")
		asnDBPathFlag    = flag.String("asnDB", "", "Optional: Path to a MaxMind/DB-IP ASN database (.mmdb)")
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

	// Load ASN database if provided
	var asnDB *maxminddb.Reader
	if *asnDBPathFlag != "" {
		var err error
		asnDB, err = maxminddb.Open(*asnDBPathFlag)
		if err != nil {
			slog.Error("Failed to open ASN database", "path", *asnDBPathFlag, "error", err)
			os.Exit(1)
		}
		defer asnDB.Close()
		slog.Info("Using ASN database", "path", *asnDBPathFlag)
	}

	// Load SOAX config from environment variables
	cfg, err := soax.NewConfig(
		os.Getenv("SOAX_API_KEY"),
		os.Getenv("SOAX_PACKAGE_KEY"),
		os.Getenv("SOAX_PACKAGE_ID"),
		os.Getenv("SOAX_PROXY_HOST"),
		os.Getenv("SOAX_PROXY_PORT"),
	)
	if err != nil {
		slog.Error("Failed to initialize SOAX config from environment", "error", err)
		os.Exit(1)
	}

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
	reportDir := filepath.Join(workspaceDir, "ispreport")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		slog.Error("Failed to create report directory", "path", reportDir, "error", err)
		os.Exit(1)
	}
	outputFilename := filepath.Join(reportDir, fmt.Sprintf("results-%s-countries%d.csv", sanitizedDomain, len(countries)))
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
			"domain", "country_code", "country_name", "isp", "asn", "exit_node_ip", "exit_node_isp", "discovered_ip",
			"ip_match", "geodb_asn", "geodb_as_name", "asn_match", "ech_grease",
			"go_error", "curl_exit_code", "curl_error_name", "curl_error_message",
			"dns_lookup_ms", "tcp_connection_ms", "tls_handshake_ms", "server_time_ms", "total_time_ms",
			"http_status", "http_connect_status",
		}
		if err := csvWriter.Write(header); err != nil {
			slog.Error("Failed to write CSV header", "error", err)
		}

		for r := range resultsCh {
			record := []string{
				r.Domain, r.Country, r.CountryName, r.ISP, r.ASN, r.ExitNodeIP, r.ExitNodeISP, r.DiscoveredIP,
				r.IPMatch, r.GeoDBASN, r.GeoDBASName, r.ASNMatch, strconv.FormatBool(r.ECHGrease),
				r.GoError, strconv.Itoa(r.CurlExitCode), r.CurlErrorName, r.CurlErrorMessage,
				strconv.FormatInt(r.DNSLookup.Milliseconds(), 10),
				strconv.FormatInt(r.TCPConnection.Milliseconds(), 10),
				strconv.FormatInt(r.TLSHandshake.Milliseconds(), 10),
				strconv.FormatInt(r.ServerTime.Milliseconds(), 10),
				strconv.FormatInt(r.TotalTime.Milliseconds(), 10),
				strconv.Itoa(r.HTTPStatus),
				strconv.Itoa(r.HTTPConnectStatus),
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

		isps, err := soax.ListISPs(cfg, country.Code)
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

				proxyURL := soax.BuildWebProxyURL(cfg, c.Code, isp, sid)
				slog.Debug("Testing ISP", "country", c.Code, "isp", isp, "ech_grease", ech, "session", sid)
				resultsCh <- runSoaxTest(curlPath, domain, c.Code, c.Name, isp, proxyURL, ech, *maxTimeFlag, *ipCheckURLFlag, asnDB)
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
	auditFilename := filepath.Join(reportDir, "isps-audit.json")
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
