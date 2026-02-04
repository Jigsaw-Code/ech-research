# SOAX ECH GREASE Report Generation

This tool tests ECH GREASE compatibility by issuing requests via SOAX proxies.
It iterates through a list of countries and ISPs, running tests with and without
ECH GREASE to simulate diverse network vantage points.

## Requirements

You need to build the ECH-enabled `curl` and place it in the workspace directory. See [instructions](../curl/README.md).

You also need a SOAX configuration file (`soax/cred.json` in the workspace) and a list of ISO country codes.

### Configuration File Examples

**SOAX Credentials (`soax/cred.json`)**

The SOAX configuration file should be a JSON file with the following structure:

```json
{
  "api_key": "YOUR_API_KEY",
  "package_key": "YOUR_PACKAGE_KEY",
  "package_id": "YOUR_PACKAGE_ID",
  "proxy_host": "proxy.soax.com",
  "proxy_port": 5000
}
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

To run the tool, use the `go run` command from the project root directory:

```sh
go run ./soaxreport --countries workspace/countries.csv --targetDomain www.google.com
```

This will:

1. Load the SOAX credentials (`./workspace/soax/cred.json` by default) and country list.
2. For each country, fetch the list of available ISPs.
3. For each ISP, issue requests to the target domain via a SOAX proxy, once with ECH GREASE and once without.
4. Save the results to `./workspace/soax-results-<domain>-countries<N>.csv`.

### Parameters

* `-workspace <path>`: Directory to store intermediate files. Defaults to `./workspace`.
* `-soax <path>`: Path to SOAX config JSON. Defaults to `./workspace/soax/cred.json`.
* `-countries <path>`: Path to CSV file containing country names and ISO codes (required).
* `-targetDomain <domain>`: Target domain to test. Defaults to `www.google.com`.
* `-parallelism <number>`: Maximum number of parallel requests. Defaults to `16`.
* `-verbose`: Enable verbose logging.
* `-maxTime <duration>`: Maximum time per curl request. Defaults to `30s`.
* `-curl <path>`: Path to the ECH-enabled curl binary. Defaults to `./workspace/output/bin/curl`.

### Output Format

The tool generates a CSV file (`workspace/soax-results-<domain>-countries<N>.csv`) with the following columns:

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
