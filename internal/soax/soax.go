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

package soax

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"golang.getoutline.org/sdk/x/soax"
)

// Config holds the credentials and endpoint configuration for the SOAX service.
type Config struct {
	APIKey     string `json:"api_key"`
	PackageKey string `json:"package_key"`
	PackageID  string `json:"package_id"`
	ProxyHost  string `json:"proxy_host"`
	ProxyPort  int    `json:"proxy_port"`
}

// Client provides methods to interact with the SOAX API and generate proxy configurations.
type Client struct {
	cfg *Config
	sdk *soax.Client
}

// LoadConfig reads the SOAX configuration from a JSON file.
// If ProxyHost or ProxyPort are missing in the config, default values are used.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config json: %w", err)
	}
	if cfg.ProxyHost == "" {
		cfg.ProxyHost = "proxy.soax.com"
	}
	if cfg.ProxyPort == 0 {
		cfg.ProxyPort = 5000
	}

	return &cfg, nil
}

// NewClient creates a new SOAX Client with the given configuration.
func NewClient(cfg *Config) *Client {
	return &Client{
		cfg: cfg,
		sdk: &soax.Client{
			APIKey:     cfg.APIKey,
			PackageKey: cfg.PackageKey,
		},
	}
}

// ListISPs retrieves a list of available ISP operators for the specified country code.
// countryISO should be a 2-letter ISO country code (e.g., "US").
func (c *Client) ListISPs(countryISO string) ([]string, error) {
	ctx := context.Background()
	ispMap := make(map[string]bool)

	if mIsps, err := c.sdk.GetMobileISPs(ctx, countryISO, "", ""); err == nil {
		for _, isp := range mIsps {
			ispMap[isp] = true
		}
	}

	if rIsps, err := c.sdk.GetResidentialISPs(ctx, countryISO, "", ""); err == nil {
		for _, isp := range rIsps {
			ispMap[isp] = true
		}
	}

	if len(ispMap) == 0 {
		return nil, fmt.Errorf("no ISPs found for country %s", countryISO)
	}

	var isps []string
	for isp := range ispMap {
		isps = append(isps, isp)
	}
	return isps, nil
}

// BuildWebProxyURL constructs an authenticated HTTPS proxy URL for a specific country and ISP.
// An optional sessionID can be provided for sticky sessions; if empty, a random one is generated.
func (c *Client) BuildWebProxyURL(countryISO, ispName, sessionID string) string {
	if sessionID == "" {
		sessionID = generateRandomString(10)
	}

	params := []string{"package", c.cfg.PackageID}
	if countryISO != "" {
		params = append(params, "country", strings.ToLower(countryISO))
	}
	if ispName != "" {
		params = append(params, "isp", strings.ToLower(ispName))
	}
	if sessionID != "" {
		params = append(params, "sessionid", sessionID)
	}
	params = append(params, "sessionlength", "300")

	u := &url.URL{
		Scheme: "https",
		User:   url.UserPassword(strings.Join(params, "-"), c.cfg.PackageKey),
		Host:   fmt.Sprintf("%s:%d", c.cfg.ProxyHost, c.cfg.ProxyPort),
	}

	return u.String()
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
