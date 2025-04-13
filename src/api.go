package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/denwerk/moco/src/types"
)

func makeRequest(cfg *Config, method, path string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/%s", cfg.MocoDomain, path)
	LogAPIRequest(method, url, body)

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		LogAPIError(err)
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+cfg.MocoAPIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogAPIError(err)
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogAPIError(err)
		return nil, err
	}

	LogAPIResponse(resp.StatusCode, responseBody)

	if resp.StatusCode >= 400 {
		// Try to parse error message from response
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(responseBody, &errorResponse); err == nil && errorResponse.Error != "" {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorResponse.Error)
		}
		// If we can't parse the error message, return the raw response
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func fetchProjects(cfg *Config) ([]types.Project, error) {
	body, err := makeRequest(cfg, "GET", "projects/assigned", nil)
	if err != nil {
		return nil, err
	}

	var projects []types.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		LogAPIError(err)
		return nil, fmt.Errorf("error unmarshaling projects: %v", err)
	}

	return projects, nil
}

func fetchTimeEntries(cfg *Config, date string) ([]types.TimeEntry, error) {
	endDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		LogAPIError(err)
		return nil, err
	}

	startDate := endDate.AddDate(0, 0, -6)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	path := fmt.Sprintf("activities?from=%s&to=%s", startDateStr, endDateStr)
	body, err := makeRequest(cfg, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var entries []types.TimeEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		LogAPIError(err)
		return nil, fmt.Errorf("error unmarshaling time entries: %v", err)
	}

	return entries, nil
}

func submitTimeEntry(cfg *Config, entry types.TimeEntry) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		LogAPIError(err)
		return err
	}

	_, err = makeRequest(cfg, "POST", "activities", jsonData)
	return err
}

func deleteTimeEntry(cfg *Config, id int) error {
	path := fmt.Sprintf("activities/%d", id)
	_, err := makeRequest(cfg, "DELETE", path, nil)
	return err
}
