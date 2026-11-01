package source

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Popularity comes from Homebrew's public analytics: how many times each
// formula and cask was installed in the last year across all Homebrew users.
// It is a rough proxy for "is this a well-known tool", nothing more.
//
// The datasets are several megabytes, so they are fetched in the background,
// cached on disk, and never block the interface.

const (
	formulaeAnalyticsURL = "https://formulae.brew.sh/api/analytics/install/365d.json"
	casksAnalyticsURL    = "https://formulae.brew.sh/api/analytics/cask-install/365d.json"
	popularityCacheDays  = 14
)

type popularityCache struct {
	Fetched int64          `json:"fetched"`
	Counts  map[string]int `json:"counts"`
}

type analyticsDocument struct {
	Items []struct {
		Formula string `json:"formula"`
		Cask    string `json:"cask"`
		Count   string `json:"count"`
	} `json:"items"`
}

func popularityCachePath() (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "dronagiri", "popularity.json"), nil
}

// CachedPopularity returns the stored counts when they are still fresh.
func CachedPopularity() (map[string]int, bool) {
	path, err := popularityCachePath()
	if err != nil {
		return nil, false
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var cache popularityCache
	if err := json.Unmarshal(raw, &cache); err != nil {
		return nil, false
	}

	age := time.Since(time.Unix(cache.Fetched, 0))
	if age > popularityCacheDays*24*time.Hour {
		return cache.Counts, false
	}
	return cache.Counts, true
}

// FetchPopularity downloads both datasets and stores them for next time.
func FetchPopularity(timeout time.Duration) (map[string]int, error) {
	client := &http.Client{Timeout: timeout}
	counts := make(map[string]int)

	for _, url := range []string{formulaeAnalyticsURL, casksAnalyticsURL} {
		if err := fetchAnalytics(client, url, counts); err != nil {
			return nil, err
		}
	}

	savePopularity(counts)
	return counts, nil
}

func fetchAnalytics(client *http.Client, url string, into map[string]int) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "dronagiri")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("analytics: %s", response.Status)
	}

	var document analyticsDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return err
	}

	for _, item := range document.Items {
		name := item.Formula
		if name == "" {
			name = item.Cask
		}
		if name == "" {
			continue
		}
		count, err := strconv.Atoi(strings.ReplaceAll(item.Count, ",", ""))
		if err != nil {
			continue
		}
		into[name] = count
	}
	return nil
}

func savePopularity(counts map[string]int) {
	path, err := popularityCachePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	raw, err := json.Marshal(popularityCache{Fetched: time.Now().Unix(), Counts: counts})
	if err != nil {
		return
	}
	_ = os.WriteFile(path, raw, 0o644)
}
