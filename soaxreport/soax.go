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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SoaxConfig holds the credentials and endpoint configuration for the SOAX service.
type SoaxConfig struct {
	APIKey     string `json:"api_key"`
	PackageKey string `json:"package_key"`
	PackageID  string `json:"package_id"`
	ProxyHost  string `json:"proxy_host"`
	ProxyPort  int    `json:"proxy_port"`
}

// SoaxClient provides methods to interact with the SOAX API and generate proxy configurations.
type SoaxClient struct {
	Config *SoaxConfig
}

// NewSoaxClient creates a new SoaxClient by loading configuration from a JSON file.
// If ProxyHost or ProxyPort are missing in the config, default values are used.
func NewSoaxClient(configPath string) (*SoaxClient, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg SoaxConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config json: %w", err)
	}
	if cfg.ProxyHost == "" {
		cfg.ProxyHost = "proxy.soax.com"
	}
	if cfg.ProxyPort == 0 {
		cfg.ProxyPort = 5000
	}

	return &SoaxClient{Config: &cfg}, nil
}

// ListISPs retrieves a list of available ISP operators for the specified country code.
// countryISO should be a 2-letter ISO country code (e.g., "US").
func (s *SoaxClient) ListISPs(countryISO string) ([]string, error) {
	url := fmt.Sprintf("https://api.soax.com/api/get-country-operators?api_key=%s&package_key=%s&country_iso=%s",
		s.Config.APIKey, s.Config.PackageKey, strings.ToLower(countryISO))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ISPs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %s, body: %s", resp.Status, string(body))
	}

	var isps []string
	if err := json.NewDecoder(resp.Body).Decode(&isps); err != nil {
		return nil, fmt.Errorf("failed to decode ISP list: %w", err)
	}
	return isps, nil
}

// BuildProxyURL constructs an authenticated HTTPS proxy URL for a specific country and ISP.
// An optional sessionID can be provided for sticky sessions; if empty, a random one is generated.
func (s *SoaxClient) BuildProxyURL(countryISO, ispName, sessionID string) string {
	if sessionID == "" {
		sessionID = generateRandomString(10)
	}

	ispName = url.QueryEscape(strings.ToLower(ispName))
	countryISO = strings.ToLower(countryISO)

	proxyUser := fmt.Sprintf("package-%s-country-%s-isp-%s-sessionid-%s-sessionlength-300",
		s.Config.PackageID, countryISO, ispName, sessionID)

	return fmt.Sprintf("https://%s:%s@%s:%d",
		proxyUser, s.Config.PackageKey, s.Config.ProxyHost, s.Config.ProxyPort)
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
