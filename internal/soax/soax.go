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
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"golang.getoutline.org/sdk/x/soax"
)

// Config holds the credentials and endpoint configuration for the SOAX service.
type Config struct {
	APIKey     string
	PackageKey string
	PackageID  string
	ProxyHost  string
	ProxyPort  int
}

// NewConfig creates a new Config, validating required fields and setting defaults.
func NewConfig(apiKey, packageKey, packageID, proxyHost, proxyPortStr string) (*Config, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if packageKey == "" {
		return nil, fmt.Errorf("package key is required")
	}
	if packageID == "" {
		return nil, fmt.Errorf("package ID is required")
	}
	if proxyHost == "" {
		proxyHost = "proxy.soax.com"
	}

	proxyPort := 5000
	if proxyPortStr != "" {
		p, err := strconv.Atoi(proxyPortStr)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy port %q: %v", proxyPortStr, err)
		}
		proxyPort = p
	}

	return &Config{
		APIKey:     apiKey,
		PackageKey: packageKey,
		PackageID:  packageID,
		ProxyHost:  proxyHost,
		ProxyPort:  proxyPort,
	}, nil
}

// ListISPs retrieves a list of available ISP operators for the specified country code.
// It combines both mobile and residential ISPs using the provided SDK client.
func ListISPs(cfg *Config, countryISO string) ([]string, error) {
	sdkClient := &soax.Client{
		APIKey:     cfg.APIKey,
		PackageKey: cfg.PackageKey,
	}

	ctx := context.Background()
	ispMap := make(map[string]bool)

	if mIsps, err := sdkClient.GetMobileISPs(ctx, countryISO, "", ""); err == nil {
		for _, isp := range mIsps {
			ispMap[isp] = true
		}
	}

	if rIsps, err := sdkClient.GetResidentialISPs(ctx, countryISO, "", ""); err == nil {
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
func BuildWebProxyURL(cfg *Config, countryISO, ispName, sessionID string) string {
	if sessionID == "" {
		sessionID = generateRandomString(10)
	}

	params := []string{"package", cfg.PackageID}
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
		User:   url.UserPassword(strings.Join(params, "-"), cfg.PackageKey),
		Host:   fmt.Sprintf("%s:%d", cfg.ProxyHost, cfg.ProxyPort),
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
