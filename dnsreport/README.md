# DNS Report Generation

This document outlines the steps to generate a report on DNS query latency and HTTPS RR feature usage. The process involves collecting data, running analysis in an interactive Jupyter notebook, and visualizing the findings.

The notebook (`dnsreport/report.ipynb`) is pre-populated with cached analysis results so that you can view the final report immediately, but you can also re-run the analysis at any time using your own collected data.

## Step 1: Setup

The analysis is packaged within an interactive Jupyter notebook that contains pre-populated results and can be dynamically re-run.

We recommend using [uv](https://docs.astral.sh/uv/) for virtual environment and dependency management (you can install it by following the [official installation guide](https://docs.astral.sh/uv/getting-started/installation/)). If you prefer standard Python tools (like `python3 -m venv`), you can use them instead.

1.  **Create and activate the virtual environment:**
    From the `ech-research` directory, run:
    ```sh
    uv venv
    source .venv/bin/activate
    ```

2.  **Install dependencies:**
    ```sh
    uv pip install -r requirements.txt
    ```

## Step 2: Collect DNS Data

From the `ech-research` folder, run the data collection tool. The following command will query the top 10,000 domains 5 times each, which is a good sample for the report.

```sh
go run ./dnsreport -topN 10000 -numQueries 5
```

This will save the results to `./workspace/results-top10000-n5.csv`.

### Parameters

* `-workspace <path>`: Directory to store intermediate files. Defaults to `./workspace`.
* `-trancoID <id>`: The ID of the Tranco list to use. Defaults to `7NZ4X`.
* `-topN <number>`: The number of top domains to analyze. Defaults to 100.
* `-parallelism <number>`: Maximum number of parallel requests. Defaults to 10.
* `-numQueries <number>`: Number of times to query each domain. Defaults to 1.

### Output Format

The tool generates a CSV file (`workspace/results-top<N>-n<M>.csv`) with the following columns:

* `domain`: The domain that was queried.
* `rank`: The rank of the domain in the Tranco list.
* `run`: The run number of the query.
* `query_type`: The type of DNS query (A, AAAA, HTTPS).
* `timestamp`: When the query was performed (RFC3339Nano).
* `duration_ms`: How long the query took in milliseconds.
* `error`: Any error that occurred during the query.
* `rcode`: The DNS response code (e.g., NoError, NXDomain).
* `cnames`: The CNAME records in the answer section, formatted as a JSON array.
* `answers`: The resource records in the answer section (excluding CNAMEs), formatted as a JSON array.
* `additionals`: The resource records in the additional section, formatted as a JSON array.

---

## Step 3: Analyze and Visualize via Jupyter Notebook

All sorting, filtering, graphing, and table compilation are carried out interactively inside the `dnsreport/report.ipynb` notebook.

You can open this notebook natively in **VS Code** (choose the `.venv` kernel in the top-right corner) or in a web browser using the standard command:

```sh
cd dnsreport
jupyter notebook report.ipynb
```

The notebook is pre-populated with default analysis results for immediate viewing, but you can also rerun the cells at any time to process your newly collected data files.

All generated plots, tables, and statistics appear directly inline. Once you have finished executing the analysis cells, you can finalize your conclusions within the notebook's markdown sections or export the entire document as a self-contained HTML/PDF report.
