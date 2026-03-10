# SOAX ECH GREASE Report Generation

This tool tests ECH GREASE compatibility by issuing requests via SOAX proxies.
It iterates through a list of countries and ISPs, running tests with and without
ECH GREASE to simulate diverse network vantage points.

## Requirements

You need to build the ECH-enabled `curl` and place it in the workspace directory. See [instructions](../curl/README.md).

You also need to set the SOAX credentials as environment variables and provide a list of ISO country codes.

### Configuration

**SOAX Credentials (Environment Variables)**

Set the following environment variables with your SOAX API details:

```bash
export SOAX_API_KEY="YOUR_API_KEY"
export SOAX_PACKAGE_KEY="YOUR_PACKAGE_KEY"
export SOAX_PACKAGE_ID="YOUR_PACKAGE_ID"
# Optional overrides:
# export SOAX_PROXY_HOST="proxy.soax.com"
# export SOAX_PROXY_PORT="5000"
```

**Country List (`countries.csv`)**

The countries file should be a CSV file containing country names and their 2-letter ISO codes. Lines starting with `#` are ignored.

```csv
"United States",US
"United Kingdom",GB
"Germany",DE
# Add more countries as needed
"Virgin Islands, U.S.",VI
```

You can download a complete list of country codes from [here](https://raw.githubusercontent.com/datasets/country-list/master/data.csv).

## Running

To run the tool, ensure your environment variables are set, then use the `go run` command from the project root directory:

```sh
go run ./soaxreport --targetDomain www.google.com
```

This will:

1. Load the SOAX credentials from the environment and the country list (`./workspace/countries.csv` by default).
2. For each country, fetch the list of available ISPs.
3. For each ISP, issue requests to the target domain via a SOAX proxy, once with ECH GREASE and once without.
4. Save the results to `./workspace/soax-results-<domain>-countries<N>.csv`.

### Parameters

* `-workspace <path>`: Directory to store intermediate files. Defaults to `./workspace`.
* `-countries <path>`: Path to CSV file containing country names and ISO codes. Defaults to `./workspace/countries.csv`.
* `-targetDomain <domain>`: Target domain to test. Defaults to `www.google.com`.
* `-parallelism <number>`: Maximum number of parallel requests. Defaults to `16`.
* `-verbose`: Enable verbose logging.
* `-maxTime <duration>`: Maximum time per curl request. Defaults to `30s`.
* `-curl <path>`: Path to the ECH-enabled curl binary. Defaults to `./workspace/output/bin/curl`.

### Output Format

The tool generates two output files in the workspace directory:

1. **Results CSV** (`workspace/soax-results-<domain>-countries<N>.csv`): Contains the detailed test results for each request.
2. **ISP Audit Log** (`workspace/soax-isps-audit.json`): A JSON file mapping each country code to the list of ISPs discovered and used during the test. This is useful for auditing coverage.

The CSV file contains the following columns:

* `domain`: The domain that was tested.
* `country_code`: The 2-letter ISO country code.
* `country_name`: The full name of the country.
* `isp`: The ISP name of the proxy used.
* `asn`: The ASN of the proxy exit node.
* `exit_node_ip`: The IP address of the proxy exit node.
* `exit_node_isp`: The ISP name reported by the proxy exit node (from headers).
* `ech_grease`: `true` if ECH GREASE was enabled for the request, `false` otherwise.
* `error`: Any error that occurred during the request.
* `curl_exit_code`: The exit code returned by the `curl` command.
* `curl_error_name`: The human-readable name corresponding to the `curl` exit code.
* `dns_lookup_ms`: The duration of the DNS lookup.
* `tcp_connection_ms`: The duration of the TCP connection.
* `tls_handshake_ms`: The duration of the TLS handshake.
* `server_time_ms`: The time from the end of the TLS handshake to the first byte of the response.
* `total_time_ms`: The total duration of the request.
* `http_status`: The HTTP status code of the response.
