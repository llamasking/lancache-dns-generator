package cachedomains

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/sync/errgroup"
)

const ENDPOINT_LATEST_COMMIT = "https://api.github.com/repos/uklans/cache-domains/commits/master"
const ENDPOINT_RAW_FILE = "https://raw.githubusercontent.com/uklans/cache-domains/master/"
const MAX_CONCURRENCY = 6

var last_known_sha = ""
var last_known_list = ""

type serviceList struct {
	CacheDomains []struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		DomainFiles  []string `json:"domain_files"`
		Notes        string   `json:"notes,omitempty"`
		MixedContent bool     `json:"mixed_content,omitempty"`
	} `json:"cache_domains"`
}

// Gets latest commit SHA from Github API.
func getLatestSHA() (string, error) {
	resp, err := http.Get(ENDPOINT_LATEST_COMMIT)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// I don't want to bother with making a struct for this
	var json_resp map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&json_resp); err != nil {
		return "", err
	}

	if value, ok := json_resp["sha"].(string); value != "" && ok {
		return value, nil
	} else {
		return "", fmt.Errorf("github api did not return latest commit")
	}
}

// Gets the service data from cache_domains.json
func getServiceList() (serviceList, error) {
	var list = serviceList{}

	resp, err := http.Get(ENDPOINT_RAW_FILE + "cache_domains.json")
	if err != nil {
		return list, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&list); err != nil {
		return list, err
	}

	return list, nil
}

// Gets contents of host files with names given.
func getRawFiles(files []string) (string, error) {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(MAX_CONCURRENCY)

	ch := make(chan string, len(files))

	for _, file := range files {
		g.Go(func() error {
			req, err := http.NewRequestWithContext(ctx, "GET", ENDPOINT_RAW_FILE+file, nil)
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("github returned non-ok status for: %s", file)
			}

			if b, err := io.ReadAll(resp.Body); err != nil {
				return err
			} else {
				ch <- "# " + file + "\n" + string(b)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return "", err
	}
	close(ch)

	var full_resp = ""
	for resp := range ch {
		full_resp += resp
	}
	return full_resp, nil
}

// Gets the contents of all domain files from github, combined into one string
// The data is in the structure described in the README of https://github.com/uklans/cache-domains
func GetAllDomainFiles() (string, error) {
	// If no new commits have happened, use the cached list
	if sha, err := getLatestSHA(); err != nil {
		return "", err
	} else if sha == last_known_sha {
		return last_known_list, nil
	} else {
		last_known_sha = sha
	}

	services, err := getServiceList()
	if err != nil {
		return "", err
	}

	domain_files := []string{}
	for _, d := range services.CacheDomains {
		domain_files = append(domain_files, d.DomainFiles...)
	}

	list, err := getRawFiles(domain_files)
	if err != nil {
		return "", err
	}

	last_known_list = list
	return list, nil
}
