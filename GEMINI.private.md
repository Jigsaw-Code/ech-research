# THE DEFINITIVE RESEARCH ARCHIVE: COMPLETE PROJECT RECORD (MAXIMUM FIDELITY)

> **⚠️ ARCHIVAL NOTICE**: This document represents the cumulative knowledge of the project.
> *   **IMMUTABLE**: Do not delete or truncate.
> *   **STATUS**: Restored & Merged (Legacy + Modern + Operational).

---

# PART I: LEGACY SYSTEM (`greasereport`) - CODE ANATOMY

> *Restoring the deep-dive analysis of the legacy codebase.*

### 1.1 The Architecture of `greasereport/main.go`

The `greasereport` tool was built as a highly concurrent, low-overhead prober. Unlike Python scripts which might suffer from GIL issues during high-concurrency network I/O, this Go binary was designed to saturate the local network link with thousands of simultaneous HEAD requests.

#### A. The Imports: Engineering Decisions
```go
import (
    "encoding/csv"
    "golang.org/x/sync/semaphore"
    // ...
)
```
1.  **`encoding/csv`**: We deliberately chose this over manual string splitting/joining. Why? Because domain names (from Tranco) or Curl error messages can contain commas, quotes, or newlines. The `csv` package handles RFC 4180 escaping automatically, preventing data corruption in the results file.
2.  **`golang.org/x/sync/semaphore`**: This external dependency is critical. It provides a Weighted Semaphore. We use this to implement "Bounded Concurrency". Without it, launching 10,000 goroutines would exhaust the OS file descriptors (`ulimit -n`), causing artificial `CURLE_COULDNT_CONNECT` errors that look like blocking but are actually local resource starvation.

#### B. The Data Model: `TestResult`
The `TestResult` struct is a direct memory mapping of our research hypothesis.
```go
type TestResult struct {
    Domain        string        // The identity (e.g., "google.com")
    Rank          int           // The value (Tranco #1 vs #999,999)
    ECHGrease     bool          // The variable (True/False)
    
    // The Diagnostics
    CurlExitCode  int           // The raw signal (0-96)
    CurlErrorName string        // The human signal
    
    // The Forensics (Timings)
    DNSLookup     time.Duration // Time to resolve IP
    TCPConnection time.Duration // Time to SYN-ACK
    TLSHandshake  time.Duration // Time to ServerHello/Finished
    // ...
}
```
*   **Deep Logic**: By separating `DNSLookup`, `TCPConnection`, and `TLSHandshake`, we perform "Triangulation".
    *   **Case A**: `DNS=0`. Conclusion: Local DNS failure or UDP block.
    *   **Case B**: `DNS>0`, `TCP=0`. Conclusion: IP blocking or routing blackhole.
    *   **Case C**: `TCP>0`, `TLS=0`. Conclusion: The TCP connection was established, but the TLS handshake failed. **This is the exact signature of ECH blocking**, where a middlebox sees the ClientHello, dislikes the ECH extension, and drops the packet or sends RST.

#### C. The Exit Code Map: `curlExitCodeNames`
We statically compiled a map of 96 libcurl exit codes.
*   **Engineering Reason**: We cannot parse error strings ("Connection refused") reliably across OS languages or curl versions. Integers are the only stable contract.
*   **Key Codes for ECH**:
    *   **35 (`CURLE_SSL_CONNECT_ERROR`)**: The "Smoking Gun".
    *   **60 (`CURLE_SSL_CACERT`)**: Potential MITM attack (Middlebox presents invalid cert).

#### D. The `runTest` Engine
This function constructs the shell command.
```go
args := []string{
    "-w", "dnslookup:%{time_namelookup},tcpconnect:%{time_connect},tlsconnect:%{time_appconnect}...",
    "--head", "--ech", "grease", ...
}
```
*   **The `-w` Hack**: This is the most sophisticated part of the legacy tool. We bypass Go's network stack entirely and rely on `curl`'s internal timers.
    *   `%{time_appconnect}`: This variable is only available if the SSL handshake completes. If it's zero, we know the handshake failed.
*   **The Parsing Loop**:
    ```go
    parts := strings.Split(out.String(), ",")
    for _, part := range parts {
        // kv splitting...
    }
    ```
    This loop effectively deserializes the custom `-w` string back into struct fields.

#### E. The Concurrency Pipeline (`main`)
1.  **Initialization**:
    *   `sem := semaphore.NewWeighted(int64(*parallelismFlag))`
2.  **The Loop**:
    *   Iterate `tranco.List`.
    *   `sem.Acquire(ctx, 1)`: **Blocking Call**. If 10 requests are running, this line pauses the main thread.
    *   `go func() { defer sem.Release(1) ... }()`: Launches the worker.
3.  **The Sink**:
    *   `resultsCh` acts as a buffer. A separate goroutine writes to the CSV file. This decouples network latency (variable) from disk I/O latency (constant), ensuring the workers never block on disk writes.

---

# PART II: MODERN SYSTEM (`soaxreport`) - ARCHITECTURE & EVOLUTION

This section analyzes the current production system `soaxreport/main.go`, designed for global residential proxy testing.

### 2.1 The "Sticky Session" Breakthrough (CRITICAL)
**Problem**: Residential IPs rotate. Control and Experiment requests might hit different IPs, invalidating the test.
**Solution**: Append `;session_id={RANDOM}` to the proxy username.
**Code Analysis**:
```go
runSessionID := time.Now().Format("0102150405")
sessionID := fmt.Sprintf("%s%s%d", runSessionID, country.Code, i)
proxyURL := client.BuildProxyURL(c.Code, isp, sessionID)
```
*   **`runSessionID`**: Base entropy (timestamp).
*   **`country.Code + i`**: Ensures every ISP pair gets a *unique* session ID, but the Control/Experiment *within* that pair share the *same* ID. This guarantees they exit via the same physical device.

### 2.2 Metadata Parsing (Side-Channel)
**Challenge**: `curl` through a proxy doesn't know the exit IP.
**Mechanism**: SOAX injects headers into the `CONNECT` response.
**Code Analysis (`runSoaxTest`)**:
```go
// 1. Request Injection
ProxyHeaders: []string{"Respond-With: ip,isp,asn"}

// 2. Response Parsing (Stderr)
for line := range strings.SplitSeq(res.Stderr, "\n") {
    if strings.HasPrefix(line, "< Node-") {
        // Parse "Node-IP", "Node-ISP"
    }
}
```
*   **Why Stderr?**: `curl -v` prints headers to stderr. We parse this stream to capture the "Truth" of where we actually exited.

### 2.3 Concurrency Architecture
*   **Nested Loops**: `Countries -> ISPs -> Test Pair`.
*   **Rate Limiting**: `semaphore` (Weight 16). Tuned down from 50 to prevent SOAX gateway timeouts (Exit 28).
*   **Atomic Progress**: `total.Add(...)` and `finished.Add(1)` provide thread-safe progress logging.

---

# PART III: DATA ANALYSIS LOGIC (`analyze.py`)

### 3.1 The "Namibia" Bug
*   **Issue**: Pandas reads "NA" (Namibia) as `NaN`.
*   **Fix**: `pd.read_csv(..., keep_default_na=False)`.

### 3.2 The Grouping Heuristic
```python
grouped = df.groupby(['domain', 'country_code', 'isp'])
```
*   **Granularity**: ISP-level. Country averages mask censorship.

### 3.3 Divergence Visualization (Red/Green Bars)
**Logic**:
*   `Difference = (ECH_Success - Control_Success)`
*   **Negative (< 0)**: **Red Bar**. `Control OK, ECH Fail`. = **Blocking**.
*   **Positive (> 0)**: **Green Bar**. `Control Fail, ECH OK`. = **Bypass**.
*   **Significance**: Visualizes the "Iran Anomaly" (equal fail rates but disjoint sets) by showing both bars if they exist.

---

# PART IV: OPERATIONAL MANUAL

### 4.1 Configuration Files
**`workspace/soax/cred.json`**:
```json
{
  "username": "...",
  "password": "...",
  "proxy_address": "proxy.soax.com",
  "proxy_port": 443
}
```

**`workspace/data/countries.csv`**:
```csv
Name,Code
United States,US
China,CN
Russia,RU
```

### 4.2 Build & Run
1.  **Build Curl**: `curl/build-curl.sh`.
2.  **Build Tool**: `go build ./cmd/soaxreport`.
3.  **Execute**:
    ```bash
    ./soaxreport --countries workspace/data/countries.csv --targetDomain www.google.com --parallelism 16
    ```

---

# PART V: PROJECT HISTORY & RESEARCH LOG

### 5.1 Phase 1: Discovery
*   Verified SOAX API manually via `curl -x ...`.
*   Discovered ISO code requirement for ISP listing.

### 5.2 Phase 2: Prototype (`greasereport`)
*   Validated Tranco list logic.
*   Found local testing insufficient (100% success).

### 5.3 Phase 3: Production (`soaxreport`)
*   Implemented Sticky Sessions.
*   Implemented Metadata Parsing.
*   Collected datasets for Google, Mail, YouTube.

### 5.4 Phase 4: Analysis
*   Identified 99%+ success rates.
*   Found specific blocks in Iran, Russia, Mauritania.
*   Concluded "No Systematic Blocking".

---

# PART VI: `soaxreport/main.go` LINE-BY-LINE SOURCE ANALYSIS

This section provides an exhaustive walkthrough of the production collection tool, detailing the engineering rationale behind every block.

### 6.1 Data Structures & Error Mapping
The file starts by defining `TestResult`, an evolution of the `greasereport` struct.
*   **Fields `ExitNodeIP` & `ExitNodeISP`**: Added specifically to combat "Proxy Fraud" or "Gateway Spoofing". We don't trust the SOAX API's claim of an ISP; we verify it via the headers returned during the actual connection.
*   **`internal/curl` abstraction**: Unlike the legacy version, we now use a dedicated package to encapsulate the `curl` subprocess logic, improving testability.

### 6.2 The CSV Ingestion Engine (`loadCountries`)
```go
func loadCountries(path string) ([]Country, error) { ... }
```
*   **The `#` Support**: We added `reader.Comment = '#'` to allow the research team to comment out specific countries without deleting rows. This is essential for long-running campaigns where certain regions might be temporarily unreachable.
*   **Field Validation**: `reader.FieldsPerRecord = 2` ensures that a malformed CSV row (e.g., missing a comma) results in an immediate failure rather than silent data corruption.

### 6.3 The Core Runner (`runSoaxTest`)
This function is where the network-level experimentation happens.
*   **`ProxyHeaders` Injection**: We explicitly send `Respond-With: ip,isp,asn`. This tells the SOAX gateway to perform a lookup on the exit peer and return the results in the `CONNECT` response headers.
*   **Metadata Extraction Logic**:
    ```go
    for line := range strings.SplitSeq(res.Stderr, "\n") {
        line = strings.TrimSpace(line)
        if !strings.HasPrefix(line, "< Node-") { continue }
        // ... Split and capture ASN, IP, ISP
    }
    ```
    *Analysis*: Why parse `stderr`? Because `curl`'s `-w` option only provides timing and status codes. The HTTP headers of the proxy handshake are only available in the verbose log (`-v` / `stderr`).

### 6.4 The `main` Orchestration Logic
The entry point manages the complex lifecycle of a global campaign.

#### 1. Flag Handling
The tool supports 8 distinct flags.
*   `--maxTime`: Defaults to 30s. This is significantly longer than `greasereport` (10s) because residential proxies often take 5-10s just to establish the initial circuit.
*   `--parallelism`: Defaults to 16. Through trial and error, we found that values > 32 frequently result in SOAX gateway errors (Exit 56 - Recv Error) due to socket exhaustion.

#### 2. The Worker Pipeline
The tool uses a "One Producer, Two Consumers" pattern for each job.
*   **Producer**: The nested loop (Countries -> ISPs).
*   **ISP Discovery**: `client.ListISPs(country.Code)` is called synchronously to fetch the available pool.
*   **The Session ID Logic**:
    ```go
    sessionID := fmt.Sprintf("%s%s%d", runSessionID, country.Code, i)
    ```
    *Why?*: By including the loop index `i` in the session ID, we ensure that the Control pass and the Experiment pass for ISP #5 always use the exact same proxy session, while ISP #6 uses a different one. This isolates the variable to **just the ECH flag**.

#### 3. Thread-Safe Progress Tracking
```go
var total, finished atomic.Int32
// ...
progress := fmt.Sprintf("%d/%d", finished.Add(1), total.Load())
slog.Info("Finished", "country", c.Code, "isp", isp, "progress", progress)
```
*Analysis*: Standard integers would suffer from "lost updates" when 16 goroutines update them simultaneously. `atomic.Int32` provides a lock-free way to track progress accurately.

#### 4. The Sink (CSV Writer Goroutine)
A dedicated goroutine owns the `csv.Writer`.
*   **Buffered Writing**: It reads from `resultsCh`. If the disk is slow, the results buffer in the channel rather than blocking the network workers.
*   **Flush Policy**: `defer csvWriter.Flush()` ensures the last few bytes are written to disk before the program exits.

---
**END OF COMPLETE ARCHIVE**
