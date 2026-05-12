# ECH GREASE Compatibility Report: Top 10,000 Domains

## Methodology

The `greasereport` tool was used to evaluate the compatibility of the top 10,000 domains (from the Tranco list, ID `7NZ4X`) with Encrypted ClientHello (ECH) GREASE. 

For each domain, two HTTP `HEAD` requests were issued concurrently using a custom ECH-enabled `curl` binary:
1. **Control Run:** ECH GREASE disabled (`--ech false`).
2. **Test Run:** ECH GREASE enabled (`--ech grease`).

The HTTP status code, `curl` exit code, and various connection timings (DNS, TCP, TLS handshake, and Server Time) were recorded for each request. A maximum timeout of 10 seconds was enforced per request to handle unresponsive servers.

## Findings

A total of 10,000 domains were tested. The raw measurements are available in `grease-results-top10000.csv`.

* **Overall Strict Success Rate:** Both the control and test runs achieved an identical strict HTTP success rate (status codes 2xx/3xx) of exactly **56.71%** (5,671 domains).
* **TLS Connection Success Rate:** Application-level HTTP errors (e.g., 403 Forbidden, 404 Not Found) indicate that the TLS handshake and underlying connection were successful, but the server rejected the request at the application layer. When reclassifying these 1,602 domains as successes from a connection standpoint, the effective TLS success rate is **72.73%** (7,273 domains).
* **Failure Breakdown:** The 4,329 domains that did not return a 2xx/3xx HTTP status in the baseline run can be categorized as follows:
  * **Network & TLS Errors (2,727 domains):**
    * 1,742: `CURLE_COULDNT_RESOLVE_HOST` (DNS resolution failed)
    * 398: `CURLE_OPERATION_TIMEDOUT` (Connection timed out)
    * 375: `CURLE_SSL_CACERT` (SSL certificate verification failed)
    * 93: `CURLE_COULDNT_CONNECT` (Connection refused)
    * 92: `CURLE_SSL_CONNECT_ERROR` (Failed to negotiate TLS)
    * 27: Other Curl Errors (e.g., Receive errors)
  * **Application-Level HTTP Errors (1,602 domains):**
    * 716: `HTTP 403 Forbidden` (Common for bot-protection blocking `curl`)
    * 561: `HTTP 404 Not Found`
    * 83: `HTTP 405 Method Not Allowed`
    * 82: `HTTP 400 Bad Request`
    * 30: `HTTP 503 Service Unavailable`
    * 24: `HTTP 429 Too Many Requests` (Rate limiting)
    * 106: Other HTTP Errors (e.g., 401, 500, 502)

* **Initial Anomalies:** An initial algorithmic analysis of the results identified exactly 10 domains that succeeded in the control run but failed when ECH GREASE was enabled. 

### Re-verification of Anomalous Domains

To distinguish true ECH GREASE interference from transient network errors (such as temporary rate limiting, dynamic routing failures, or intermittent timeouts), the 10 anomalous domains were subjected to manual re-testing.

The following domains were re-ran:
1. `telemetry.mozilla.org`
2. `tradedoubler.com`
3. `www.microsoft.com`
4. `vmailru.net`
5. `rutracker.org`
6. `futurecdn.net`
7. `pt.m.wikipedia.org`
8. `login.caixa.gov.br`
9. `bergfex.at`
10. `statcounter.com`

**Results of Re-verification:**

* **Consistent Success:** 7 domains (`telemetry.mozilla.org`, `tradedoubler.com`, `www.microsoft.com`, `vmailru.net`, `futurecdn.net`, `pt.m.wikipedia.org`, `statcounter.com`) succeeded on subsequent requests regardless of whether ECH GREASE was enabled or disabled. This indicates the initial failures were transient network blips.
* **Consistent Failure:** `rutracker.org` consistently timed out for both control and test requests, indicating standard server-side blocking or unreachability rather than ECH interference.
* **Rate Limiting / Server Behavior:** `bergfex.at` and `login.caixa.gov.br` exhibited flapping responses (e.g., alternating between HTTP `429 Too Many Requests`, HTTP `404 Not Found`, and expected `301/302` redirects) across both control and test requests. These failures are attributed to strict rate limiting and stateful server-side behavior rather than the presence of the GREASE extension.

## Conclusion

Upon rigorous verification, **no consistent ECH GREASE interference was observed** among the top 10,000 domains tested. All initial failures associated with ECH GREASE were confirmed to be transient network anomalies, server unreachability, or strict server-side rate limiting policies.
