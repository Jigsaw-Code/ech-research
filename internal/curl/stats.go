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
	"strconv"
	"strings"
	"time"
)

// Stats captures timing and status metrics from a curl execution.
type Stats struct {
	// DNSLookupTimestamp is the cumulative time from the start until the name
	// lookup is completed (time_namelookup).
	DNSLookupTimestamp time.Duration

	// TCPConnectTimestamp is the cumulative time from the start until the TCP
	// connection is completed (time_connect).
	TCPConnectTimestamp time.Duration

	// TLSConnectTimestamp is the cumulative time from the start until the
	// SSL/TLS handshake is completed (time_appconnect).
	TLSConnectTimestamp time.Duration

	// ServerResponseTimestamp is the cumulative time from the start until the
	// first byte is received (time_starttransfer).
	ServerResponseTimestamp time.Duration

	// TotalTimeTimestamp is the total time from the start until the operation is
	// fully completed (time_total).
	TotalTimeTimestamp time.Duration

	// HTTPStatus is the HTTP response code (http_code).
	HTTPStatus int
}

const (
	// statsPrefix is the delimiter used to identify the statistics block in the output.
	statsPrefix = "\n|||CURL_STATS|||\t"

	// statsFormat is the format string passed to curl's -w flag.
	statsFormat = statsPrefix +
		"dnslookup:%{time_namelookup}," +
		"tcpconnect:%{time_connect}," +
		"tlsconnect:%{time_appconnect}," +
		"servertime:%{time_starttransfer}," +
		"total:%{time_total}," +
		"httpstatus:%{http_code}"
)

// parseStats extracts Stats from the curl output by looking for the statsPrefix.
func parseStats(stdout string) Stats {
	var s Stats
	idx := strings.LastIndex(stdout, statsPrefix)
	if idx == -1 {
		return s
	}

	raw := strings.TrimSpace(stdout[idx+len(statsPrefix):])
	for part := range strings.SplitSeq(raw, ",") {
		kv := strings.Split(part, ":")
		if len(kv) != 2 {
			continue
		}

		key, val := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		switch key {
		case "dnslookup":
			s.DNSLookupTimestamp = parseDuration(val)
		case "tcpconnect":
			s.TCPConnectTimestamp = parseDuration(val)
		case "tlsconnect":
			s.TLSConnectTimestamp = parseDuration(val)
		case "servertime":
			s.ServerResponseTimestamp = parseDuration(val)
		case "total":
			s.TotalTimeTimestamp = parseDuration(val)
		case "httpstatus":
			s.HTTPStatus, _ = strconv.Atoi(val)
		}
	}
	return s
}

// parseDuration converts a seconds-based float string to time.Duration.
func parseDuration(s string) time.Duration {
	f, _ := strconv.ParseFloat(s, 64)
	return time.Duration(f * float64(time.Second))
}
