# ECH GREASE Report Generation

This tool, located in the `greasereport/` directory, issues HEAD requests to a list of domains with and without ECH GREASE to test for compatibility. It uses a custom-built ECH-enabled `curl` binary for these requests.

## Requirements

You need to build the ECH-enabled `curl` and place it in the workspace directory. See [instructions](../curl/README.md).

## Running

To run the tool, use the `go run` command from the `ech-test` directory:

```sh
go run ./greasereport --topN 100
```

This will:

1. Create a `./workspace` directory if it doesn't exist.
2. Download the Tranco top 1 million domains list (if not already present).
3. Issue HEAD requests to the top 100 domains, once with ECH GREASE and once without.
4. Save the results to `./workspace/grease-results-top100.csv`.

### Parameters

* `-workspace <path>`: Directory to store intermediate files. Defaults to `./workspace`.
* `-trancoID <id>`: The ID of the Tranco list to use. Defaults to `7NZ4X`.
* `-topN <number>`: The number of top domains to analyze. Defaults to 100.
* `-parallelism <number>`: Maximum number of parallel requests. Defaults to 10.
* `-curl <path>`: Path to the ECH-enabled curl binary. Defaults to `./workspace/output/bin/curl`.
* `-maxTime <duration>`: Maximum time per curl request. Defaults to `10s`.

### Output Format

The tool generates a CSV file (`workspace/grease-results-top<N>.csv`) with the following columns:

* `domain`: The domain that was tested.
* `rank`: The rank of the domain in the Tranco list.
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

## Report

After running the `greasereport` tool, a `report` subdirectory is created within the `greasereport` directory. This directory contains:

*   `report.md`: A summary of the ECH GREASE connectivity analysis.
*   `analyze.py`: The Python script used for the analysis.
*   `grease-results-top<N>.csv`: The raw data from the test run.
