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

package echtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type TestResult struct {
	Domain    string
	ECHGrease bool

	GoError          string
	CurlExitCode     int
	CurlErrorName    string
	CurlErrorMessage string

	DNSLookup     time.Duration
	TCPConnection time.Duration
	TLSHandshake  time.Duration
	ServerTime    time.Duration
	TotalTime     time.Duration

	HTTPStatus        int
	HTTPConnectStatus int

	Stderr string
}

// curlExitCodeNames maps curl exit codes to their CURL_* string representations.
var curlExitCodeNames = map[int]string{
	1:  "CURLE_UNSUPPORTED_PROTOCOL",
	2:  "CURLE_FAILED_INIT",
	3:  "CURLE_URL_MALFORMAT",
	4:  "CURLE_NOT_BUILT_IN",
	5:  "CURLE_COULDNT_RESOLVE_PROXY",
	6:  "CURLE_COULDNT_RESOLVE_HOST",
	7:  "CURLE_COULDNT_CONNECT",
	8:  "CURLE_WEIRD_SERVER_REPLY",
	9:  "CURLE_REMOTE_ACCESS_DENIED",
	11: "CURLE_FTP_WEIRD_PASV_REPLY",
	13: "CURLE_FTP_WEIRD_227_FORMAT",
	14: "CURLE_FTP_CANT_GET_HOST",
	15: "CURLE_FTP_CANT_RECONNECT",
	17: "CURLE_FTP_COULDNT_SET_TYPE",
	18: "CURLE_PARTIAL_FILE",
	19: "CURLE_FTP_COULDNT_RETR_FILE",
	21: "CURLE_QUOTE_ERROR",
	22: "CURLE_HTTP_RETURNED_ERROR",
	23: "CURLE_WRITE_ERROR",
	25: "CURLE_UPLOAD_FAILED",
	26: "CURLE_READ_ERROR",
	27: "CURLE_OUT_OF_MEMORY",
	28: "CURLE_OPERATION_TIMEDOUT",
	30: "CURLE_FTP_PORT_FAILED",
	31: "CURLE_FTP_COULDNT_USE_REST",
	33: "CURLE_RANGE_ERROR",
	34: "CURLE_HTTP_POST_ERROR",
	35: "CURLE_SSL_CONNECT_ERROR",
	36: "CURLE_BAD_DOWNLOAD_RESUME",
	37: "CURLE_FILE_COULDNT_READ_FILE",
	38: "CURLE_LDAP_CANNOT_BIND",
	39: "CURLE_LDAP_SEARCH_FAILED",
	41: "CURLE_FUNCTION_NOT_FOUND",
	42: "CURLE_ABORTED_BY_CALLBACK",
	43: "CURLE_BAD_FUNCTION_ARGUMENT",
	45: "CURLE_INTERFACE_FAILED",
	47: "CURLE_TOO_MANY_REDIRECTS",
	48: "CURLE_UNKNOWN_OPTION",
	49: "CURLE_TELNET_OPTION_SYNTAX",
	51: "CURLE_PEER_FAILED_VERIFICATION",
	52: "CURLE_GOT_NOTHING",
	53: "CURLE_SSL_ENGINE_NOTFOUND",
	54: "CURLE_SSL_ENGINE_SETFAILED",
	55: "CURLE_SEND_ERROR",
	56: "CURLE_RECV_ERROR",
	58: "CURLE_SSL_CERTPROBLEM",
	59: "CURLE_SSL_CIPHER",
	60: "CURLE_SSL_CACERT",
	61: "CURLE_BAD_CONTENT_ENCODING",
	62: "CURLE_LDAP_INVALID_URL",
	63: "CURLE_FILESIZE_EXCEEDED",
	64: "CURLE_USE_SSL_FAILED",
	65: "CURLE_SEND_FAIL_REWIND",
	66: "CURLE_SSL_ENGINE_INITFAILED",
	67: "CURLE_LOGIN_DENIED",
	68: "CURLE_TFTP_NOTFOUND",
	69: "CURLE_TFTP_PERM",
	70: "CURLE_REMOTE_DISK_FULL",
	71: "CURLE_TFTP_ILLEGAL",
	72: "CURLE_TFTP_UNKNOWNID",
	73: "CURLE_REMOTE_FILE_EXISTS",
	74: "CURLE_TFTP_NOSUCHUSER",
	75: "CURLE_CONV_FAILED",
	76: "CURLE_CONV_REQD",
	77: "CURLE_SSL_CACERT_BADFILE",
	78: "CURLE_REMOTE_FILE_NOT_FOUND",
	79: "CURLE_SSH",
	80: "CURLE_SSL_SHUTDOWN_FAILED",
	81: "CURLE_AGAIN",
	82: "CURLE_SSL_CRL_BADFILE",
	83: "CURLE_SSL_ISSUER_ERROR",
	84: "CURLE_FTP_PRET_FAILED",
	85: "CURLE_RTSP_CSEQ_ERROR",
	86: "CURLE_RTSP_SESSION_ERROR",
	87: "CURLE_FTP_BAD_FILE_LIST",
	88: "CURLE_CHUNK_FAILED",
	89: "CURLE_NO_CONNECTION_AVAILABLE",
	90: "CURLE_SSL_PINNEDPUBKEYNOTMATCH",
	91: "CURLE_SSL_INVALIDCERTSTATUS",
	92: "CURLE_HTTP2_STREAM",
	93: "CURLE_RECURSIVE_API_CALL",
	94: "CURLE_AUTH_ERROR",
	95: "CURLE_HTTP3",
	96: "CURLE_QUIC_CONNECT_ERROR",
}

// Run executes a curl command against the specified domain.
func Run(
	curlPath string,
	domain string,
	echGrease bool,
	maxTime time.Duration,
	proxyURL string,
	proxyHeaders []string,
) TestResult {
	result := TestResult{
		Domain:    domain,
		ECHGrease: echGrease,
	}

	targetURL := "https://" + domain

	args := []string{
		"-w", ":::BEGIN_JSON:::%{json}:::END_JSON:::",
		"--head",
		"--max-time",
		strconv.FormatFloat(maxTime.Seconds(), 'f', -1, 64),
	}

	// Handle proxy options
	if proxyURL != "" {
		args = append(args, "--proxy", proxyURL)
		for _, h := range proxyHeaders {
			args = append(args, "--proxy-header", h)
		}
		// If using a proxy with headers, we usually need verbose mode to see the proxy response.
		// If proxy headers are provided, we assume the caller wants to read them from stderr.
		if len(proxyHeaders) > 0 {
			args = append(args, "-v")
		} else {
			args = append(args, "-s")
		}
	} else {
		args = append(args, "-s")
	}

	if echGrease {
		args = append(args, "--ech", "grease")
	} else {
		args = append(args, "--ech", "false")
	}
	args = append(args, targetURL)

	cmd := exec.Command(curlPath, args...)

	// Setup environment for custom curl (matching internal/curl/runner.go)
	binDir := filepath.Dir(curlPath)
	libDir := filepath.Join(filepath.Dir(binDir), "lib")
	if libStat, err := os.Stat(libDir); err == nil && libStat.IsDir() {
		cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+libDir)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Stderr = stderr.String() // Always capture stderr for caller

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.CurlExitCode = exitError.ExitCode()
			result.CurlErrorName = curlExitCodeNames[result.CurlExitCode]
		} else {
			result.GoError = fmt.Sprintf("failed to execute curl: %v", err)
			return result
		}
	} else {
		// Even if err is nil, there might be curl-level errors recorded in stderr
		// that the caller might be interested in, though standard execution succeeded.
	}

	// Parse JSON output
	if stdout.Len() > 0 {
		outStr := stdout.String()
		const startMarker = ":::BEGIN_JSON:::"
		const endMarker = ":::END_JSON:::"
		startIndex := strings.Index(outStr, startMarker)
		endIndex := strings.Index(outStr, endMarker)

		if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
			jsonStr := outStr[startIndex+len(startMarker) : endIndex]
			var curlOut struct {
				TimeNamelookup    float64 `json:"time_namelookup"`
				TimeConnect       float64 `json:"time_connect"`
				TimeAppconnect    float64 `json:"time_appconnect"`
				TimeStarttransfer float64 `json:"time_starttransfer"`
				TimeTotal         float64 `json:"time_total"`
				HTTPCode          int     `json:"http_code"`
				HTTPConnect       int     `json:"http_connect"`
				Errormsg          string  `json:"errormsg"`
			}
			if err := json.Unmarshal([]byte(jsonStr), &curlOut); err == nil {
				result.DNSLookup = time.Duration(curlOut.TimeNamelookup * float64(time.Second))
				result.TCPConnection = time.Duration(curlOut.TimeConnect * float64(time.Second))
				result.TLSHandshake = time.Duration(curlOut.TimeAppconnect * float64(time.Second))
				result.ServerTime = time.Duration(curlOut.TimeStarttransfer * float64(time.Second))
				result.TotalTime = time.Duration(curlOut.TimeTotal * float64(time.Second))
				result.HTTPStatus = curlOut.HTTPCode
				result.HTTPConnectStatus = curlOut.HTTPConnect
				result.CurlErrorMessage = curlOut.Errormsg
			} else {
				if result.GoError == "" {
					result.GoError = fmt.Sprintf("failed to parse curl json: %v", err)
				}
			}
		} else {
			if result.GoError == "" {
				result.GoError = "could not find JSON boundaries in curl output"
			}
		}
	}

	return result
}
