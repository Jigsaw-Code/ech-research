# SOAX ECH GREASE Report Generation

This tool tests ECH GREASE compatibility by issuing requests via SOAX proxies.
It iterates through a list of countries and ISPs, running tests with and without
ECH GREASE to simulate diverse network vantage points.

## Requirements

You need to build the ECH-enabled `curl` and place it in the workspace directory. See [instructions](../curl/README.md).

You also need to set the SOAX credentials as environment variables and provide a list of ISO country codes.

**(Optional) ASN Validation:** To independently verify the ASN of the proxy exit nodes, you can provide a local IP-to-ASN database in `.mmdb` format (such as [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) or [DB-IP ASN Lite](https://db-ip.com/db/download/ip-to-asn-lite)).

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

A default country list is provided at `ispreport/report/countries.csv`. You can copy this file to your workspace directory (`workspace/countries.csv`) before running the tool, or specify its path using the `--countries` parameter.

You can download a complete list of country codes from [here](https://raw.githubusercontent.com/datasets/country-list/master/data.csv) as well.

## Running

To run the tool, ensure your environment variables are set, then use the `go run` command from the project root directory.

**Basic Run:**

```sh
go run ./ispreport --targetDomain www.google.com
```

**With Independent ASN Validation (Recommended):**
First, download a free IP-to-ASN `.mmdb` database (e.g., from DB-IP) to your workspace.
```sh
go run ./ispreport --targetDomain www.google.com --asnDB workspace/dbip-asn-lite.mmdb
```

**With Custom IP Check URL and Verbose Logging:**
```sh
go run ./ispreport --targetDomain www.google.com --ipCheckURL https://ifconfig.me/ip --verbose
```

This will:

1. Load the SOAX credentials from the environment and the country list (`./workspace/countries.csv` by default).
2. For each country, fetch the list of available ISPs.
3. For each ISP, discover the real proxy exit IP via the `ipCheckURL`.
4. Issue requests to the target domain via the SOAX proxy, once with ECH GREASE and once without.
5. Save the results to `./workspace/ispreport/results-<domain>-countries<N>.csv`.

### Parameters

* `-workspace <path>`: Directory to store intermediate files. Defaults to `./workspace`.
* `-countries <path>`: Path to CSV file containing country names and ISO codes. Defaults to `./workspace/countries.csv`.
* `-targetDomain <domain>`: Target domain to test. Defaults to `www.google.com`.
* `-parallelism <number>`: Maximum number of parallel requests. Defaults to `16`.
* `-verbose`: Enable verbose logging.
* `-maxTime <duration>`: Maximum time per curl request. Defaults to `30s`.
* `-curl <path>`: Path to the ECH-enabled curl binary. Defaults to `./workspace/bin/curl`.
* `-ipCheckURL <url>`: URL used to discover the real external IP of the proxy. Defaults to `https://ipv4.icanhazip.com/`.
* `-asnDB <path>`: Optional path to a MaxMind or DB-IP `.mmdb` database file for independent ASN verification.

### Output Format

The tool generates two output files in the workspace directory:

1. **Results CSV** (`workspace/ispreport/results-<domain>-countries<N>.csv`): Contains the detailed test results for each request.
2. **ISP Audit Log** (`workspace/ispreport/isps-audit.json`): A JSON file mapping each country code to the list of ISPs discovered and used during the test. This is useful for auditing coverage.

The CSV file contains the following columns:

* `domain`: The domain that was tested.
* `country_code`: The 2-letter ISO country code.
* `country_name`: The full name of the country.
* `isp`: The ISP name of the proxy used.
* `asn`: The ASN of the proxy exit node as reported by the SOAX proxy headers.
* `exit_node_isp`: The ISP name reported by the SOAX proxy headers.
* `geodb_asn`: The ASN corresponding to the `discovered_ip`, looked up in the `-asnDB` (if provided).
* `geodb_as_name`: The AS organization name corresponding to the `discovered_ip`, looked up in the `-asnDB` (if provided).
* `asn_match`: `true` if the SOAX-reported `asn` matches the `geodb_asn`, `false` otherwise.
* `ech_grease`: `true` if ECH GREASE was enabled for the request, `false` otherwise.
* `go_error`: Any error that occurred during the request.
* `curl_exit_code`: The exit code returned by the `curl` command.
* `curl_error_name`: The human-readable name corresponding to the `curl` exit code.
* `curl_error_message`: The detailed error message from curl (if available).
* `dns_lookup_ms`: The duration of the DNS lookup.
* `tcp_connection_ms`: The duration of the TCP connection.
* `tls_handshake_ms`: The duration of the TLS handshake.
* `server_time_ms`: The time from the end of the TLS handshake to the first byte of the response.
* `total_time_ms`: The total duration of the request.
* `http_status`: The HTTP status code of the response.
* `http_connect_status`: The HTTP status code from the proxy connection.

## Generating the Final Report

The visual report notebook (`ispreport/report/report.ipynb`) is pre-populated with cached multi-country analysis results and plots for domains like `www.google.com` and `cloudflare-ech.com`. You can view and explore the final report immediately without setting up any SOAX proxy accounts or running any data collection.

If you collect new proxy data, you can easily re-run the notebook to update the visualizations.

### 1. Organize the Data

The notebook expects data to be organized in subdirectories within `ispreport/report/` named after the tested domain.

1. Create a subdirectory for your results (e.g., for `www.google.com`):
   ```bash
   mkdir -p ispreport/report/www_google_com
   ```

2. Copy and rename the generated results from the `workspace` directory:
   ```bash
   # Use a wildcard to match the generated file with the number of countries
   cp workspace/ispreport/results-www_google_com-countries*.csv ispreport/report/www_google_com/results.csv
   cp workspace/ispreport/isps-audit.json ispreport/report/www_google_com/isps-audit.json
   ```

### 2. Setup the Environment

Running the notebook requires Python 3 and several data analysis libraries.

We recommend using [uv](https://docs.astral.sh/uv/) for virtual environment and dependency management (you can install it by following the [official installation guide](https://docs.astral.sh/uv/getting-started/installation/)). If you prefer standard Python tools (like `python3 -m venv`), you can use them instead.

```bash
# From the project root, create and activate the environment:
uv venv
source .venv/bin/activate

# Install required dependencies:
uv pip install -r requirements.txt
```

### 3. Run the Notebook

You can open this notebook natively in **VS Code** (choose the `.venv` kernel in the top-right corner) or in a web browser using the standard command:

```bash
cd ispreport/report
jupyter notebook report.ipynb
```

1. In the first code cell of the notebook, update the `DOMAIN` variable to match the name of the subdirectory you created (e.g., `DOMAIN = "www_google_com"`).
2. Run the cells in the notebook to generate the analysis and visualizations.
