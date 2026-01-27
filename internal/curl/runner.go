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

package curl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Runner handles the execution of a specific curl binary.
// It manages environment setup (e.g., LD_LIBRARY_PATH) required for custom builds.
type Runner struct {
	curlPath string
	libPath  string
}

// Args defines the execution parameters for a single curl run.
type Args struct {
	// Proxy specifies the proxy URL to use.
	// Corresponds to the "--proxy" flag.
	Proxy string

	// ProxyHeaders specifies custom headers to send to the proxy.
	// Each string should be in "Key: Value" format.
	// Corresponds to the "--proxy-header" flag.
	ProxyHeaders []string

	// ECH specifies the Encrypted ClientHello mode.
	// Corresponds to the "--ech" flag.
	// Use ECHGrease, ECHTrue, ECHFalse constants.
	// If empty (ECHNone), the flag is omitted.
	ECH ECHMode

	// Verbose enables verbose output.
	// If true, adds "-v".
	// If false, adds "-s" (silent mode) by default to keep output clean.
	Verbose bool

	// Timeout sets the maximum time allowed for the transfer.
	// Corresponds to the "--max-time" flag.
	// If 0, no timeout is set.
	Timeout time.Duration

	// MeasureStats enables capturing performance metrics using curl's -w flag.
	// If true, Stats will be populated in the Result.
	MeasureStats bool
}

// ECHMode defines the available Encrypted ClientHello modes for curl.
type ECHMode string

const (
	// ECHGrease enables ECH GREASE mode ("--ech grease").
	ECHGrease ECHMode = "grease"
	// ECHTrue enables ECH ("--ech true").
	ECHTrue ECHMode = "true"
	// ECHFalse disables ECH ("--ech false").
	ECHFalse ECHMode = "false"
	// ECHNone indicates that the --ech flag should not be sent.
	ECHNone ECHMode = ""
)

// Result represents the raw outcome of a curl execution.
type Result struct {
	// ExitCode is the exit status of the curl process.
	// 0 indicates success. See ExitCodeName for error name mapping.
	ExitCode int

	// Stdout contains the standard output of the curl command.
	Stdout string

	// Stderr contains the standard error of the curl command.
	// In verbose mode, this contains debug information and headers.
	Stderr string

	// Stats contains performance metrics if MeasureStats was enabled.
	Stats Stats
}

// NewRunner creates a new Runner for the specified curl binary.
// It automatically detects the associated library path (bin/curl -> lib/)
// to ensure shared libraries are found.
func NewRunner(curlPath string) *Runner {
	r := &Runner{curlPath: curlPath}

	binDir := filepath.Dir(curlPath)
	libDir := filepath.Join(filepath.Dir(binDir), "lib")
	if libStat, err := os.Stat(libDir); err == nil && libStat.IsDir() {
		r.libPath = libDir
	}

	return r
}

// Run executes curl with the provided arguments and returns the result.
func (r *Runner) Run(url string, args Args) (*Result, error) {
	var cmdArgs []string

	if args.Verbose {
		cmdArgs = append(cmdArgs, "-v")
	} else {
		cmdArgs = append(cmdArgs, "-s")
	}

	if args.Timeout > 0 {
		cmdArgs = append(cmdArgs, "--max-time", strconv.FormatFloat(args.Timeout.Seconds(), 'f', -1, 64))
	}

	if args.Proxy != "" {
		cmdArgs = append(cmdArgs, "--proxy", args.Proxy)
	}

	for _, h := range args.ProxyHeaders {
		cmdArgs = append(cmdArgs, "--proxy-header", h)
	}

	if args.ECH != ECHNone {
		cmdArgs = append(cmdArgs, "--ech", string(args.ECH))
	}

	if args.MeasureStats {
		cmdArgs = append(cmdArgs, "-w", statsFormat)
	}

	cmdArgs = append(cmdArgs, url)
	cmd := exec.Command(r.curlPath, cmdArgs...)
	if r.libPath != "" {
		cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+r.libPath)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	result := &Result{}
	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if args.MeasureStats {
		result.Stats = parseStats(result.Stdout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("failed to execute curl: %w", err)
		}
	}

	return result, nil
}
