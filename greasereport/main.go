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
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Jigsaw-Code/ech-research/internal/echtest"
	"github.com/Jigsaw-Code/ech-research/internal/tranco"
	"github.com/Jigsaw-Code/ech-research/internal/workspace"
	"golang.org/x/sync/semaphore"
)

type DomainTestResult struct {
	Rank int
	echtest.TestResult
}

func main() {
	var (
		workspaceFlag   = flag.String("workspace", "./workspace", "Directory to store intermediate files")
		trancoIDFlag    = flag.String("trancoID", "7NZ4X", "Tranco list ID to use")
		topNFlag        = flag.Int("topN", 100, "Number of top domains to analyze")
		parallelismFlag = flag.Int("parallelism", 10, "Maximum number of parallel requests")
		verboseFlag     = flag.Bool("verbose", false, "Enable verbose logging")
		maxTimeFlag     = flag.Duration("maxTime", 10*time.Second, "Maximum time per curl request")
		curlPathFlag    = flag.String("curl", "", "Path to the ECH-enabled curl binary")
	)
	flag.Parse()

	if *verboseFlag {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	}

	// Set up workspace directory.
	workspaceDir := workspace.EnsureWorkspace(*workspaceFlag)

	// Determine curl binary path.
	curlPath := *curlPathFlag
	if curlPath == "" {
		curlPath = filepath.Join(workspaceDir, "bin", "curl")
	}

	// Ensure Tranco list is present.
	trancoList, err := tranco.NewTrancoList(workspaceDir, *trancoIDFlag)
	if err != nil {
		slog.Error("Failed to get Tranco list", "error", err)
		os.Exit(1)
	}

	// Read top N domains from Tranco CSV.
	domains, err := trancoList.TopDomains(*topNFlag)
	if err != nil {
		slog.Error("Failed to read domains from Tranco CSV", "error", err)
		os.Exit(1)
	}

	// Create new output CSV file.
	outputFilename := filepath.Join(workspaceDir, fmt.Sprintf("grease-results-top%d.csv", *topNFlag))
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		slog.Error("Failed to create output CSV file", "path", outputFilename, "error", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	csvWriter := csv.NewWriter(outputFile)
	defer csvWriter.Flush()

	header := []string{"domain", "rank", "ech_grease", "error", "curl_exit_code", "curl_error_name", "dns_lookup_ms", "tcp_connection_ms", "tls_handshake_ms", "server_time_ms", "total_time_ms", "http_status"}
	if err := csvWriter.Write(header); err != nil {
		slog.Error("Failed to write CSV header", "error", err)
		os.Exit(1)
	}

	resultsCh := make(chan DomainTestResult, 2*(*topNFlag))

	var csvWg sync.WaitGroup
	csvWg.Add(1)
	go func() {
		defer csvWg.Done()
		for result := range resultsCh {
			errorStr := result.GoError
			if errorStr == "" && result.CurlErrorMessage != "" {
				errorStr = result.CurlErrorMessage
			} else if errorStr == "" && result.Stderr != "" && result.CurlExitCode != 0 {
				errorStr = result.Stderr
			}

			record := []string{
				result.Domain,
				strconv.Itoa(result.Rank),
				strconv.FormatBool(result.ECHGrease),
				errorStr,
				strconv.Itoa(result.CurlExitCode),
				result.CurlErrorName,
				strconv.FormatInt(result.DNSLookup.Milliseconds(), 10),
				strconv.FormatInt(result.TCPConnection.Milliseconds(), 10),
				strconv.FormatInt(result.TLSHandshake.Milliseconds(), 10),
				strconv.FormatInt(result.ServerTime.Milliseconds(), 10),
				strconv.FormatInt(result.TotalTime.Milliseconds(), 10),
				strconv.Itoa(result.HTTPStatus),
			}
			if err := csvWriter.Write(record); err != nil {
				slog.Error("Failed to write record to CSV", "error", err)
			}
		}
	}()

	sem := semaphore.NewWeighted(int64(*parallelismFlag))
	var wg sync.WaitGroup

	for _, domain := range domains {
		wg.Add(2)
		if err := sem.Acquire(context.Background(), 1); err != nil {
			slog.Error("Failed to acquire semaphore", "domain", domain.Name, "error", err)
			continue
		}
		go func(d tranco.Domain) {
			defer sem.Release(1)
			defer wg.Done()
			slog.Info("Testing domain", "rank", d.Rank, "domain", d.Name, "ech_grease", false)
			res := echtest.Run(curlPath, d.Name, false, *maxTimeFlag, "", nil)
			resultsCh <- DomainTestResult{Rank: d.Rank, TestResult: res}
		}(domain)

		if err := sem.Acquire(context.Background(), 1); err != nil {
			slog.Error("Failed to acquire semaphore", "domain", domain.Name, "error", err)
			continue
		}
		go func(d tranco.Domain) {
			defer sem.Release(1)
			defer wg.Done()
			slog.Info("Testing domain", "rank", d.Rank, "domain", d.Name, "ech_grease", true)
			res := echtest.Run(curlPath, d.Name, true, *maxTimeFlag, "", nil)
			resultsCh <- DomainTestResult{Rank: d.Rank, TestResult: res}
		}(domain)
	}

	wg.Wait()
	close(resultsCh)

	csvWg.Wait()

	slog.Info("Done. Results saved to", "path", outputFilename)
}
