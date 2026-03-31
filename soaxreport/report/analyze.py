#!/usr/bin/env python3
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from datetime import datetime
import sys
import os
import argparse

sns.set_theme(style="whitegrid")

def main():
    parser = argparse.ArgumentParser(description="Analyze SOAX ECH test results.")
    parser.add_argument("csv_file", help="Input CSV file from soaxreport")
    parser.add_argument("--output-dir", "-o", default=".", help="Directory to save the report and charts")
    args = parser.parse_args()

    input_file = args.csv_file
    output_dir = args.output_dir
    
    if not os.path.exists(output_dir):
        os.makedirs(output_dir, exist_ok=True)
        
    report_file = os.path.join(output_dir, "raw-report.md")

    try:
        # CRITICAL: keep_default_na=False is needed because 'NA' is the country code for Namibia
        # but pandas interprets it as NaN by default.
        df = pd.read_csv(input_file, keep_default_na=False, na_values=[''])
    except Exception as e:
        print(f"Error reading CSV: {e}")
        sys.exit(1)
        
    df.sort_values(by=['country_code', 'isp', 'domain'], inplace=True)

    # Process pairs of Control (No ECH) and Experiment (ECH GREASE)
    results = []
    grouped = df.groupby(['domain', 'country_code', 'isp'])

    for (domain, country, isp), group in grouped:
        # Normalize types
        group = group.copy()
        group['ech_grease'] = group['ech_grease'].astype(str).str.lower().map({'true': True, 'false': False})
        
        no_ech = group[group['ech_grease'] == False]
        ech_grease = group[group['ech_grease'] == True]
        
        # Check for incomplete pairs or duplicate data
        if no_ech.empty or ech_grease.empty:
            continue
            
        c = no_ech.iloc[0]
        e = ech_grease.iloc[0]
        
        c_exit = int(c['curl_exit_code'])
        e_exit = int(e['curl_exit_code'])
        c_status = int(c['http_status'])
        e_status = int(e['http_status'])
        
        c_tls = c['tls_handshake_ms']
        e_tls = e['tls_handshake_ms']
        tls_delta = e_tls - c_tls
        
        # Detected potential network interference
        interference = (c_exit == 0) and (e_exit != 0)
        status_mismatch = (c_exit == 0) and (e_exit == 0) and (c_status != e_status)
        
        country_name = c['country_name'] if 'country_name' in c else country
        display_country = f"{country_name} ({country})"

        results.append({
            'domain': domain,
            'country_code': country,
            'country_name': country_name,
            'display_country': display_country,
            'isp': isp,
            'c_exit': c_exit,
            'e_exit': e_exit,
            'c_tls': c_tls,
            'e_tls': e_tls,
            'delta': tls_delta,
            'interference': interference,
            'status_mismatch': status_mismatch,
            'error': e['curl_error_name']
        })

    if not results:
        print("No comparable ECH Grease/No ECH pairs found.")
        sys.exit(0)

    res_df = pd.DataFrame(results)
    target_domain = res_df['domain'].iloc[0]

    # Global Overview Chart (By ISP Pair)
    healthy_pairs = 0
    potential_blocking = 0
    unreachable = 0
    
    for _, row in res_df.iterrows():
        if row['c_exit'] == 0 and row['e_exit'] == 0:
            healthy_pairs += 1
        elif row['c_exit'] == 0 and row['e_exit'] != 0:
            potential_blocking += 1
        else:
            unreachable += 1

    plt.figure(figsize=(10, 8))
    plt.pie([healthy_pairs, potential_blocking, unreachable], 
            labels=['Healthy (Both OK)', 'Potential ECH Blocking', 'Unreachable (Both Failed)'],
            autopct='%1.1f%%', 
            colors=['#66b3ff', '#ff9999', '#ffcc99'],
            startangle=90)
    plt.title(f'Global ECH Connectivity Overview\nDomain: {target_domain}', fontsize=14)
    plt.tight_layout()
    overview_plot = os.path.join(output_dir, "global_overview.png")
    plt.savefig(overview_plot, bbox_inches='tight')
    plt.close()

    # Divergence Chart (Standard TLS vs ECH GREASE ISP)
    problematic_list = []
    for _, row in res_df.iterrows():
        c_success = (row['c_exit'] == 0)
        e_success = (row['e_exit'] == 0)
        
        # Include any pair where the outcome differs
        if c_success != e_success:
            diff = (100 if e_success else 0) - (100 if c_success else 0)
            outcome_type = 'ECH Failed (Standard TLS OK)' if diff < 0 else 'ECH Succeeded (Standard TLS Failed)'
            
            problematic_list.append({
                'ISP_Label': f"{row['display_country']} | {row['isp']}",
                'Difference': diff,
                'Outcome': outcome_type
            })

    if problematic_list:
        prob_df = pd.DataFrame(problematic_list)
        # Sort by difference to group similar outcomes
        prob_df.sort_values('Difference', ascending=False, inplace=True)
        
        # Limit to top 40 for readability if there are many
        if len(prob_df) > 40:
            prob_df = prob_df.head(40)
            
        plt.figure(figsize=(12, len(prob_df) * 0.5 + 2))
        
        palette = {'ECH Failed (Standard TLS OK)': '#d62728', 'ECH Succeeded (Standard TLS Failed)': '#2ca02c'}
        
        sns.barplot(data=prob_df, x='Difference', y='ISP_Label', hue='Outcome', dodge=False, palette=palette)
        
        plt.title(f'ISP-Level Connectivity Divergence\n(Red = ECH Failed, Green = ECH Succeeded)')
        plt.xlim(-110, 110)
        plt.axvline(x=0, color='black', linewidth=0.8)
        plt.xlabel('Impact (Negative = ECH Failed, Positive = ECH Succeeded)')
        plt.ylabel('Country | ISP')
        plt.legend(loc='lower right')
        
        plt.xticks([-100, 0, 100], ['ECH FAIL', 'NEUTRAL', 'ECH SUCCESS'])
        
        drilldown_plot = os.path.join(output_dir, "problematic_countries.png")
        plt.savefig(drilldown_plot, bbox_inches='tight')
        plt.close()
    else:
        drilldown_plot = None

    # Latency Impact Distribution Chart
    plt.figure(figsize=(10, 6))
    sns.histplot(res_df['delta'], kde=True, color='teal')
    plt.axvline(x=0, color='red', linestyle='--')
    plt.title(f'TLS Handshake Latency Delta (ECH GREASE - No ECH)\nDomain: {target_domain}')
    plt.xlabel('Delta (ms)')
    plt.ylabel('Frequency')
    latency_plot = os.path.join(output_dir, "latency_delta.png")
    plt.savefig(latency_plot, bbox_inches='tight')
    plt.close()

    interferences = res_df[res_df['interference'] == True]
    avg_delta = res_df['delta'].mean()
    total_pairs = len(res_df)
    interference_count = len(interferences)
    interference_rate = (interference_count / total_pairs) * 100

    # Generate Markdown Report
    with open(report_file, 'w') as f:
        f.write("# Raw Report: ECH GREASE Connectivity (From Different Countries)\n\n")
        f.write(f"**Date:** {datetime.now().strftime('%B %d, %Y')}\\\n")
        f.write(f"**Target Domain:** `{target_domain}`\\\n")
        f.write(f"**Analyzed File:** `{os.path.basename(input_file)}`\n\n")

        f.write("## Executive Summary\n\n")
        if interference_rate < 1.0:
            conclusion = "ECH GREASE does **not** appear to cause systematic connectivity breakage."
        elif interference_rate < 5.0:
            conclusion = "ECH GREASE shows **minor** regional connectivity issues."
        else:
            conclusion = "ECH GREASE shows **significant** connectivity issues suggesting active interference."
            
        f.write(f"This report analyzed **{total_pairs}** valid ISP pairs. {conclusion}\n\n")
        f.write(f"*   **Total ISP Pairs:** {total_pairs}\n")
        f.write(f"*   **Potential Blocking:** {interference_count} ({interference_rate:.2f}%)\n")
        f.write(f"*   **Avg Latency Impact:** {avg_delta:.2f} ms\n\n")

        if (total_pairs * 2) != len(df):
            f.write(f"> ⚠️ **Data Integrity Warning:** Analyzed {total_pairs * 2} rows but file contains {len(df)} rows. Some rows were excluded because they could not be matched into a pair or were duplicates.\n\n")

        f.write("## 1. Overall Connectivity Results\n\n")
        f.write("![Global Overview](global_overview.png)\n\n")
        f.write("**Figure 1: Global Connectivity Distribution.** ")
        f.write("This chart illustrates the health of tested ISP vantage points. ")
        f.write("\"Healthy\" represents successful connections with and without ECH. ")
        f.write("\"Potential ECH Blocking\" identifies cases where only the standard TLS succeeded. ")
        f.write("\"Unreachable\" indicates ISPs that failed both tests, likely due to proxy or local network issues unrelated to ECH.\n\n")
        
        if drilldown_plot:
             f.write("## 2. Divergent ISP Connectivity (Deep Dive)\n\n")
             f.write("![Problematic Countries](problematic_countries.png)\n\n")
             f.write("**Figure 2: ISP-Level Connectivity Divergence.** ")
             f.write("This chart highlights ISPs where the connectivity outcome of ECH GREASE differs from standard TLS. ")
             f.write("**Red bars (Left)** indicate **ECH Failed** (Standard TLS worked, but ECH failed). ")
             f.write("**Green bars (Right)** indicate **ECH Succeeded** (Standard TLS failed, but ECH succeeded), showing cases where ECH maintained connectivity despite standard TLS issues.\n\n")
        
        f.write("## 3. Performance Impact\n\n")
        f.write("![Latency Delta](latency_delta.png)\n\n")
        f.write("**Figure 3: Latency Delta Distribution.** ")
        f.write("The delta is calculated as `Handshake(GREASE) - Handshake(No ECH)`. ")
        f.write("Most values clustering around 0ms suggest that ECH GREASE does not introduce significant overhead when connections succeed.\n\n")

        if not interferences.empty:
            f.write("## 4. Deep Dive: Potential Blocking\n\n")
            f.write(f"We detected **{len(interferences)}** instances where ECH GREASE failed while the control succeeded. ")
            f.write("These cases warrant further investigation to distinguish between transient network errors and active blocking.\n\n")
            affected_countries = interferences['display_country'].value_counts()
            f.write("**Affected Countries:**\n")
            for country, count in affected_countries.items():
                f.write(f"*   {country}: {count} instance(s)\n")
            f.write("\n(See Appendix A for the full list of failures)\n\n")
        else:
            f.write("## 4. Deep Dive\n\n")
            f.write("No instances of potential blocking were detected in this dataset.\n\n")

        f.write("## 5. Limitations\n\n")
        f.write("*   **Transient Errors:** Single-pass testing cannot distinguish between flaky networks and deterministic blocking. Re-runs are required for confirmation.\n")
        f.write("*   **Proxy Stability:** Residential proxies (SOAX) can be inherently unstable or slow, which may contribute to timeouts independent of ECH.\n")
        f.write("*   **Sample Size:** The number of ISPs tested per country depends on SOAX's available pool at the time of testing.\n\n")

        f.write("## Appendix A: Detailed Failure List\n\n")
        if not interferences.empty:
            f.write("| Country | ISP | No ECH Exit | GREASE Exit | Error Name |\n")
            f.write("| :--- | :--- | :--- | :--- | :--- |\n")
            for _, r in interferences.iterrows():
                f.write(f"| {r['display_country']} | {r['isp']} | {r['c_exit']} | {r['e_exit']} | {r['error']} |\n")
        else:
            f.write("_No specific ECH failures detected._\n")

        f.write("\n## Appendix B: Significant Latency Increases (>500ms)\n\n")
        slow_grease = res_df[res_df['delta'] > 500].sort_values('delta', ascending=False)
        if not slow_grease.empty:
            f.write("| Country | ISP | No ECH TLS (ms) | GREASE TLS (ms) | Delta (ms) |\n")
            f.write("| :--- | :--- | :--- | :--- | :--- |\n")
            for _, r in slow_grease.iterrows():
                f.write(f"| {r['display_country']} | {r['isp']} | {r['c_tls']:.0f} | {r['e_tls']:.0f} | {r['delta']:.0f} |\n")
        else:
            f.write("_No significant latency increases detected._\n")
        f.write("\n")

    print(f"Analysis complete. Report: {report_file}")

if __name__ == "__main__":
    main()
