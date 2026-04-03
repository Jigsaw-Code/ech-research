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

You can download a complete list of country codes from [here](https://raw.githubusercontent.com/datasets/country-list/master/data.csv).

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
5. Save the results to `./workspace/soax-results-<domain>-countries<N>.csv`.

### Parameters

* `-workspace <path>`: Directory to store intermediate files. Defaults to `./workspace`.
* `-countries <path>`: Path to CSV file containing country names and ISO codes. Defaults to `./workspace/countries.csv`.
* `-targetDomain <domain>`: Target domain to test. Defaults to `www.google.com`.
* `-parallelism <number>`: Maximum number of parallel requests. Defaults to `16`.
* `-verbose`: Enable verbose logging.
* `-maxTime <duration>`: Maximum time per curl request. Defaults to `30s`.
* `-curl <path>`: Path to the ECH-enabled curl binary. Defaults to `./workspace/output/bin/curl`.
* `-ipCheckURL <url>`: URL used to discover the real external IP of the proxy. Defaults to `https://ipv4.icanhazip.com/`.
* `-asnDB <path>`: Optional path to a MaxMind or DB-IP `.mmdb` database file for independent ASN verification.

### Output Format

The tool generates two output files in the workspace directory:

1. **Results CSV** (`workspace/soax-results-<domain>-countries<N>.csv`): Contains the detailed test results for each request.
2. **ISP Audit Log** (`workspace/soax-isps-audit.json`): A JSON file mapping each country code to the list of ISPs discovered and used during the test. This is useful for auditing coverage.

The CSV file contains the following columns:

* `domain`: The domain that was tested.
* `country_code`: The 2-letter ISO country code.
* `country_name`: The full name of the country.
* `isp`: The ISP name of the proxy used.
* `asn`: The ASN of the proxy exit node as reported by the SOAX proxy headers.
* `exit_node_ip`: The IP address of the proxy exit node as reported by the SOAX proxy headers.
* `exit_node_isp`: The ISP name reported by the SOAX proxy headers.
* `discovered_ip`: The actual public IP address of the proxy, discovered by querying `ipCheckURL`.
* `ip_match`: `true` if `exit_node_ip` equals `discovered_ip`, `false` otherwise.
* `geodb_asn`: The ASN corresponding to the `discovered_ip`, looked up in the `-asnDB` (if provided).
* `geodb_as_name`: The AS organization name corresponding to the `discovered_ip`, looked up in the `-asnDB` (if provided).
* `asn_match`: `true` if the SOAX-reported `asn` matches the `geodb_asn`, `false` otherwise.
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

## Generating the Final Report

After running the data collection tool, you can generate a visual report using the provided Jupyter notebook.

### 1. Organize the Data

The notebook expects data to be organized in subdirectories within `ispreport/report/` named after the tested domain.

1. Create a subdirectory for your results (e.g., for `www.google.com`):
   ```bash
   mkdir -p ispreport/report/www_google_com
   ```

2. Copy and rename the generated results from the `workspace` directory:
   ```bash
   # Use the actual filename generated in your workspace
   cp workspace/soax-results-www_google_com-countriesN.csv ispreport/report/www_google_com/results.csv
   cp workspace/soax-isps-audit.json ispreport/report/www_google_com/isps-audit.json
   ```

### 2. Setup the Environment

Running the notebook requires Python 3 and several data analysis libraries.

```bash
# From the project root:
# 1. Create the virtual environment if it doesn't exist
python3 -m venv workspace/.venv

# 2. Activate the virtual environment
source workspace/.venv/bin/activate

# 3. Install required dependencies
pip install pandas numpy matplotlib seaborn ipywidgets jupyter
```

### 3. Run the Notebook

1. Navigate to the report directory and start Jupyter:
   ```bash
   cd ispreport/report
   # If you didn't activate the venv yet, run: source ../../workspace/.venv/bin/activate
   jupyter notebook report.ipynb
   ```

2. In the first code cell of the notebook, update the `DOMAIN` variable to match the name of the subdirectory you created (e.g., `DOMAIN = "www_google_com"`).

3. Run all cells in the notebook to generate the analysis and visualizations.
